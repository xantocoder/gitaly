package transactions

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitaly/internal/prometheus/metrics"
)

var ErrNotFound = errors.New("transaction not found")

// Manager handles reference transactions for Praefect. It is required in order
// for Praefect to handle transactions directly instead of having to reach out
// to reference transaction RPCs.
type Manager struct {
	txIDGenerator TransactionIDGenerator
	lock          sync.Mutex
	transactions  map[uint64]*transaction
	counterMetric *prometheus.CounterVec
	delayMetric   metrics.HistogramVec
}

// TransactionIDGenerator is an interface for types that can generate transaction IDs.
type TransactionIDGenerator interface {
	// ID generates a new transaction identifier
	ID() uint64
}

type transactionIDGenerator struct {
	rand *rand.Rand
}

func newTransactionIDGenerator() *transactionIDGenerator {
	var seed [8]byte

	// Ignore any errors. In case we weren't able to generate a seed, the
	// best we can do is to just use the all-zero seed.
	cryptorand.Read(seed[:])
	source := rand.NewSource(int64(binary.LittleEndian.Uint64(seed[:])))

	return &transactionIDGenerator{
		rand: rand.New(source),
	}
}

func (t *transactionIDGenerator) ID() uint64 {
	return rand.Uint64()
}

// ManagerOpt is a self referential option for Manager
type ManagerOpt func(*Manager)

// WithCounterMetric is an option to set the counter Prometheus metric
func WithCounterMetric(counterMetric *prometheus.CounterVec) ManagerOpt {
	return func(mgr *Manager) {
		mgr.counterMetric = counterMetric
	}
}

// WithDelayMetric is an option to set the delay Prometheus metric
func WithDelayMetric(delayMetric metrics.HistogramVec) ManagerOpt {
	return func(mgr *Manager) {
		mgr.delayMetric = delayMetric
	}
}

// WithTransactionIDGenerator is an option to set the transaction ID generator
func WithTransactionIDGenerator(generator TransactionIDGenerator) ManagerOpt {
	return func(mgr *Manager) {
		mgr.txIDGenerator = generator
	}
}

// NewManager creates a new transactions Manager.
func NewManager(opts ...ManagerOpt) *Manager {
	mgr := &Manager{
		txIDGenerator: newTransactionIDGenerator(),
		transactions:  make(map[uint64]*transaction),
		counterMetric: prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"action"}),
		delayMetric:   prometheus.NewHistogramVec(prometheus.HistogramOpts{}, []string{"action"}),
	}

	for _, opt := range opts {
		opt(mgr)
	}

	return mgr
}

func (mgr *Manager) log(ctx context.Context) logrus.FieldLogger {
	return ctxlogrus.Extract(ctx).WithField("component", "transactions.Manager")
}

// CancelFunc is the transaction cancellation function returned by
// `RegisterTransaction`. Calling it will cause the transaction to be removed
// from the transaction manager.
type CancelFunc func()

// RegisterTransaction registers a new reference transaction for a set of nodes
// taking part in the transaction.
func (mgr *Manager) RegisterTransaction(ctx context.Context, nodes []string) (uint64, CancelFunc, error) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	// Use a random transaction ID. Using monotonic incrementing counters
	// that reset on restart of Praefect would be suboptimal, as the chance
	// for collisions is a lot higher in case Praefect restarts when Gitaly
	// nodes still have in-flight transactions.
	transactionID := mgr.txIDGenerator.ID()

	transaction, err := newTransaction(nodes)
	if err != nil {
		return 0, nil, err
	}

	if _, ok := mgr.transactions[transactionID]; ok {
		return 0, nil, errors.New("transaction exists already")
	}
	mgr.transactions[transactionID] = transaction

	mgr.log(ctx).WithFields(logrus.Fields{
		"transaction_id": transactionID,
		"nodes":          nodes,
	}).Debug("RegisterTransaction")

	mgr.counterMetric.WithLabelValues("registered").Inc()

	return transactionID, func() {
		mgr.cancelTransaction(transactionID, transaction)
	}, nil
}

func (mgr *Manager) cancelTransaction(transactionID uint64, transaction *transaction) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	delete(mgr.transactions, transactionID)
	transaction.cancel()
}

func (mgr *Manager) voteTransaction(ctx context.Context, transactionID uint64, node string, hash []byte) error {
	mgr.lock.Lock()
	transaction, ok := mgr.transactions[transactionID]
	mgr.lock.Unlock()

	if !ok {
		return ErrNotFound
	}

	if err := transaction.vote(node, hash); err != nil {
		return err
	}

	if err := transaction.collectVotes(ctx); err != nil {
		return err
	}

	return nil
}

// VoteTransaction is called by a client who's casting a vote on a reference
// transaction. As we currently only have primary nodes which perform reference
// transactions, this function doesn't yet do anything of interest but will
// always instruct the node to commit, if given valid transaction parameters.
// In future, it will wait for all clients of a given transaction to start the
// transaction and perform a vote.
func (mgr *Manager) VoteTransaction(ctx context.Context, transactionID uint64, node string, hash []byte) error {
	start := time.Now()
	defer func() {
		delay := time.Since(start)
		mgr.delayMetric.WithLabelValues("vote").Observe(delay.Seconds())
	}()

	mgr.counterMetric.WithLabelValues("started").Inc()

	mgr.log(ctx).WithFields(logrus.Fields{
		"transaction_id": transactionID,
		"node":           node,
		"hash":           hex.EncodeToString(hash),
	}).Debug("VoteTransaction")

	if err := mgr.voteTransaction(ctx, transactionID, node, hash); err != nil {
		mgr.log(ctx).WithFields(logrus.Fields{
			"transaction_id": transactionID,
			"node":           node,
			"hash":           hex.EncodeToString(hash),
		}).WithError(err).Error("VoteTransaction: vote failed")

		if errors.Is(err, ErrTransactionVoteFailed) {
			mgr.counterMetric.WithLabelValues("aborted").Inc()
		} else {
			mgr.counterMetric.WithLabelValues("invalid").Inc()
		}

		return err
	}

	mgr.log(ctx).WithFields(logrus.Fields{
		"transaction_id": transactionID,
		"node":           node,
		"hash":           hex.EncodeToString(hash),
	}).Debug("VoteTransaction: transaction committed")

	mgr.counterMetric.WithLabelValues("committed").Inc()

	return nil
}