package datastore

import (
	"context"
	"fmt"
	"sync"
)

// MemoryRepositoryStore is an in-memory implementation of RepositoryStore.
// Refer to the interface for method documentation.
type MemoryRepositoryStore struct {
	m sync.Mutex

	storages
	virtualStorageState
	storageState
}

type storages map[string][]string

func (s storages) secondaries(virtualStorage, primary string) ([]string, error) {
	storages, err := s.storages(virtualStorage)
	if err != nil {
		return nil, err
	}

	primaryFound := false
	secondaries := make([]string, 0, len(storages)-1)
	for _, storage := range storages {
		if storage == primary {
			primaryFound = true
			continue
		}

		secondaries = append(secondaries, storage)
	}

	if !primaryFound {
		return nil, fmt.Errorf("primary not found: %q", primary)
	}

	return secondaries, nil
}

func (s storages) storages(virtualStorage string) ([]string, error) {
	storages, ok := s[virtualStorage]
	if !ok {
		return nil, fmt.Errorf("unknown virtual storage: %q", virtualStorage)
	}

	return storages, nil
}

// virtualStorageStates represents the virtual storage's view of which repositories should exist.
// It's structured as virtual-storage->relative_path.
type virtualStorageState map[string]map[string]struct{}

// storageState contains individual storage's repository states.
// It structured as virtual-storage->relative_path->storage->generation.
type storageState map[string]map[string]map[string]int

// NewMemoryRepositoryStore returns an in-memory implementation of RepositoryStore.
func NewMemoryRepositoryStore(configuredStorages map[string][]string) *MemoryRepositoryStore {
	return &MemoryRepositoryStore{
		storages:            storages(configuredStorages),
		storageState:        make(storageState),
		virtualStorageState: make(virtualStorageState),
	}
}

func (m *MemoryRepositoryStore) GetGeneration(ctx context.Context, virtualStorage, relativePath, storage string) (int, error) {
	m.m.Lock()
	defer m.m.Unlock()

	return m.getStorageGeneration(virtualStorage, relativePath, storage), nil
}

func (m *MemoryRepositoryStore) IncrementGeneration(ctx context.Context, virtualStorage, relativePath, primary string, secondaries []string) error {
	m.m.Lock()
	defer m.m.Unlock()

	baseGen := m.getRepositoryGeneration(virtualStorage, relativePath)
	nextGen := baseGen + 1

	m.setGeneration(virtualStorage, relativePath, primary, nextGen)

	// If a secondary does not have a generation, it's in an undefined state. We'll only
	// pick secondaries on the same generation as the primary to ensure they begin from the
	// same starting state.
	if baseGen != GenerationUnknown {
		for _, secondary := range secondaries {
			currentGen := m.getStorageGeneration(virtualStorage, relativePath, secondary)
			// If the secondary is not on the same generation as the primary, the secondary
			// has failed a concurrent reference transaction. We won't increment its
			// generation as it has not applied writes in previous genereations, leaving
			// its state undefined.
			if currentGen != baseGen {
				continue
			}

			m.setGeneration(virtualStorage, relativePath, secondary, nextGen)
		}
	}

	return nil
}

func (m *MemoryRepositoryStore) SetGeneration(ctx context.Context, virtualStorage, relativePath, storage string, generation int) error {
	m.m.Lock()
	defer m.m.Unlock()

	m.setGeneration(virtualStorage, relativePath, storage, generation)

	return nil
}

func (m *MemoryRepositoryStore) DeleteRepository(ctx context.Context, virtualStorage, relativePath, storage string) error {
	m.m.Lock()
	defer m.m.Unlock()

	latestGen := m.getRepositoryGeneration(virtualStorage, relativePath)
	storageGen := m.getStorageGeneration(virtualStorage, relativePath, storage)

	m.deleteRepository(virtualStorage, relativePath)
	m.deleteStorageRepository(virtualStorage, relativePath, storage)

	if latestGen == GenerationUnknown && storageGen == GenerationUnknown {
		return RepositoryNotExistsError{
			virtualStorage: virtualStorage,
			relativePath:   relativePath,
			storage:        storage,
		}
	}

	return nil
}

func (m *MemoryRepositoryStore) RenameRepository(ctx context.Context, virtualStorage, relativePath, storage, newRelativePath string) error {
	m.m.Lock()
	defer m.m.Unlock()

	latestGen := m.getRepositoryGeneration(virtualStorage, relativePath)
	storageGen := m.getStorageGeneration(virtualStorage, relativePath, storage)

	if latestGen != GenerationUnknown {
		m.deleteRepository(virtualStorage, relativePath)
		m.createRepository(virtualStorage, newRelativePath)
	}

	if storageGen != GenerationUnknown {
		m.deleteStorageRepository(virtualStorage, relativePath, storage)
		m.setStorageGeneration(virtualStorage, newRelativePath, storage, storageGen)
	}

	if latestGen == GenerationUnknown && storageGen == GenerationUnknown {
		return RepositoryNotExistsError{
			virtualStorage: virtualStorage,
			relativePath:   relativePath,
			storage:        storage,
		}
	}

	return nil
}

func (m *MemoryRepositoryStore) GetReplicatedGeneration(ctx context.Context, virtualStorage, relativePath, source, target string) (int, error) {
	m.m.Lock()
	defer m.m.Unlock()

	sourceGeneration := m.getStorageGeneration(virtualStorage, relativePath, source)
	targetGeneration := m.getStorageGeneration(virtualStorage, relativePath, target)

	if targetGeneration != GenerationUnknown && targetGeneration >= sourceGeneration {
		return 0, DowngradeAttemptedError{
			VirtualStorage:      virtualStorage,
			RelativePath:        relativePath,
			Storage:             target,
			CurrentGeneration:   targetGeneration,
			AttemptedGeneration: sourceGeneration,
		}
	}

	return sourceGeneration, nil
}

func (m *MemoryRepositoryStore) GetConsistentSecondaries(ctx context.Context, virtualStorage, relativePath, primary string) (map[string]struct{}, error) {
	m.m.Lock()
	defer m.m.Unlock()

	secondaries, err := m.storages.secondaries(virtualStorage, primary)
	if err != nil {
		return nil, err
	}

	expectedGen := m.getRepositoryGeneration(virtualStorage, relativePath)
	if expectedGen == GenerationUnknown {
		return nil, nil
	}

	consistentSecondaries := make(map[string]struct{}, len(secondaries))
	for _, secondary := range secondaries {
		gen := m.getStorageGeneration(virtualStorage, relativePath, secondary)
		if gen == expectedGen {
			consistentSecondaries[secondary] = struct{}{}
		}
	}

	return consistentSecondaries, nil
}

func (m *MemoryRepositoryStore) IsLatestGeneration(ctx context.Context, virtualStorage, relativePath, storage string) (bool, error) {
	expected := m.getRepositoryGeneration(virtualStorage, relativePath)
	if expected == GenerationUnknown {
		return true, nil
	}

	actual := m.getStorageGeneration(virtualStorage, relativePath, storage)
	return expected == actual, nil
}

func (m *MemoryRepositoryStore) RepositoryExists(ctx context.Context, virtualStorage, relativePath string) (bool, error) {
	return m.getRepositoryGeneration(virtualStorage, relativePath) != GenerationUnknown, nil
}

func (m *MemoryRepositoryStore) GetOutdatedRepositories(ctx context.Context, virtualStorage string) (map[string]map[string]int, error) {
	m.m.Lock()
	defer m.m.Unlock()

	storages, ok := m.storages[virtualStorage]
	if !ok {
		return nil, fmt.Errorf("unknown virtual storage: %q", virtualStorage)
	}

	outdatedRepos := make(map[string]map[string]int)
	repositories, ok := m.virtualStorageState[virtualStorage]
	if !ok {
		return outdatedRepos, nil
	}

	for relativePath := range repositories {
		latestGeneration := m.getRepositoryGeneration(virtualStorage, relativePath)

		for _, storage := range storages {
			if gen := m.getStorageGeneration(virtualStorage, relativePath, storage); gen < latestGeneration {
				if outdatedRepos[relativePath] == nil {
					outdatedRepos[relativePath] = make(map[string]int)
				}

				outdatedRepos[relativePath][storage] = latestGeneration - gen
			}
		}
	}

	return outdatedRepos, nil
}

func (m *MemoryRepositoryStore) CountReadOnlyRepositories(ctx context.Context, vsPrimaries map[string]string) (map[string]int, error) {
	m.m.Lock()
	defer m.m.Unlock()

	vsReadOnly := make(map[string]int, len(vsPrimaries))
	for vs, primary := range vsPrimaries {
		vsReadOnly[vs] = 0
		relativePaths := m.virtualStorageState[vs]
		for relativePath := range relativePaths {
			expectedGeneration := m.getRepositoryGeneration(vs, relativePath)
			actualGeneration := m.getStorageGeneration(vs, relativePath, primary)
			if actualGeneration < expectedGeneration {
				vsReadOnly[vs]++
			}
		}
	}

	return vsReadOnly, nil
}

func (m *MemoryRepositoryStore) getRepositoryGeneration(virtualStorage, relativePath string) int {
	generations, ok := m.storageState[virtualStorage][relativePath]
	if !ok {
		return GenerationUnknown
	}

	max := GenerationUnknown
	for _, generation := range generations {
		if generation > max {
			max = generation
		}
	}

	return max
}

func (m *MemoryRepositoryStore) getStorageGeneration(virtualStorage, relativePath, storage string) int {
	gen, ok := m.storageState[virtualStorage][relativePath][storage]
	if !ok {
		return GenerationUnknown
	}

	return gen
}

func (m *MemoryRepositoryStore) deleteRepository(virtualStorage, relativePath string) {
	rels := m.virtualStorageState[virtualStorage]
	if rels == nil {
		return
	}

	delete(rels, relativePath)
	if len(rels) == 0 {
		delete(m.virtualStorageState, virtualStorage)
	}
}

func (m *MemoryRepositoryStore) deleteStorageRepository(virtualStorage, relativePath, storage string) {
	storages := m.storageState[virtualStorage][relativePath]
	if storages == nil {
		return
	}

	delete(storages, storage)
	if len(m.storageState[virtualStorage][relativePath]) == 0 {
		delete(m.storageState[virtualStorage], relativePath)
	}

	if len(m.storageState[virtualStorage]) == 0 {
		delete(m.storageState, virtualStorage)
	}
}

func (m *MemoryRepositoryStore) setGeneration(virtualStorage, relativePath, storage string, generation int) {
	m.createRepository(virtualStorage, relativePath)
	m.setStorageGeneration(virtualStorage, relativePath, storage, generation)
}

func (m *MemoryRepositoryStore) createRepository(virtualStorage, relativePath string) {
	if m.virtualStorageState[virtualStorage] == nil {
		m.virtualStorageState[virtualStorage] = make(map[string]struct{})
	}

	m.virtualStorageState[virtualStorage][relativePath] = struct{}{}
}

func (m *MemoryRepositoryStore) setStorageGeneration(virtualStorage, relativePath, storage string, generation int) {
	if m.storageState[virtualStorage] == nil {
		m.storageState[virtualStorage] = make(map[string]map[string]int)
	}

	if m.storageState[virtualStorage][relativePath] == nil {
		m.storageState[virtualStorage][relativePath] = make(map[string]int)
	}

	m.storageState[virtualStorage][relativePath][storage] = generation
}
