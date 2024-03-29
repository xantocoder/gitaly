package featureflag

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/metadata"
)

var (
	flagChecks = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitaly_feature_flag_checks_total",
			Help: "Number of enabled/disabled checks for Gitaly server side feature flags",
		},
		[]string{"flag", "enabled"},
	)
)

func init() {
	prometheus.MustRegister(flagChecks)
}

// IsEnabled checks if the feature flag is enabled for the passed context.
// Only returns true if the metadata for the feature flag is set to "true"
func IsEnabled(ctx context.Context, flag FeatureFlag) bool {
	val, ok := getFlagVal(ctx, flag.Name)
	if !ok {
		return flag.OnByDefault
	}

	enabled := val == "true"

	flagChecks.WithLabelValues(flag.Name, strconv.FormatBool(enabled)).Inc()

	return enabled
}

func getFlagVal(ctx context.Context, flag string) (string, bool) {
	if flag == "" {
		return "", false
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}

	val, ok := md[HeaderKey(flag)]
	if !ok {
		return "", false
	}

	if len(val) == 0 {
		return "", false
	}

	return val[0], true
}

// IsDisabled is the inverse of IsEnabled
func IsDisabled(ctx context.Context, flag FeatureFlag) bool {
	return !IsEnabled(ctx, flag)
}

const ffPrefix = "gitaly-feature-"

// HeaderKey returns the feature flag key to be used in the metadata map
func HeaderKey(flag string) string {
	return ffPrefix + strings.ReplaceAll(flag, "_", "-")
}

// AllFlags returns all feature flags with their value that use the Gitaly metadata
// prefix. Note: results will not be sorted.
func AllFlags(ctx context.Context) []string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	ffs := make([]string, 0, len(md))

	for k, v := range md {
		if !strings.HasPrefix(k, ffPrefix) {
			continue
		}
		if len(v) > 0 {
			ffs = append(ffs, fmt.Sprintf("%s:%s", strings.TrimPrefix(k, ffPrefix), v[0]))
		}
	}

	return ffs
}
