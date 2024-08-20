package inatobs

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math/rand/v2"
	"slices"

	"adamolsen.dev/buggins/internal/pkg/inatapi"
	"adamolsen.dev/buggins/internal/store"
)

type serviceConfig struct {
	projectID string
	store     *store.Queries
	pageSize  int
}

type service struct {
	serviceConfig
	api                inatapi.Api
	displayedObservers []int64
}

func newService(config serviceConfig) service {
	return service{serviceConfig: config, api: inatapi.New()}
}

func (s *service) selectUnseenObservation(
	observations []inatapi.Observation,
) (inatapi.Observation, error) {
	var (
		observationIds     []int64
		unseen             []inatapi.Observation
		seenIds            []int64
		observerMap        map[int64][]inatapi.Observation = make(map[int64][]inatapi.Observation)
		potentialObservers []int64
	)

	for _, o := range observations {
		observationIds = append(observationIds, o.ID)
	}

	seen, err := s.store.FindObservationsByIds(context.Background(), observationIds)

	if err != nil {
		return inatapi.Observation{}, fmt.Errorf("error selecting seen observations: %w", err)
	}

	for _, o := range seen {
		seenIds = append(seenIds, o.ID)
	}

	for _, o := range observations {
		if !slices.Contains(seenIds, o.ID) {
			unseen = append(unseen, o)

			items, ok := observerMap[o.UserID]

			if !ok {
				items = make([]inatapi.Observation, 0)
			}

			if ok {
				items = append(items, o)
			}

			observerMap[o.UserID] = items

			if !slices.Contains(s.displayedObservers, o.UserID) {
				potentialObservers = append(potentialObservers, o.UserID)
			}
		}
	}

	if len(unseen) <= 0 {
		return inatapi.Observation{}, errors.New("no unseen observations found")
	}

	if len(potentialObservers) <= 0 {
		potentialObservers = slices.Collect(maps.Keys(observerMap))
		s.displayedObservers = s.displayedObservers[:0]
	}

	rand.Shuffle(len(potentialObservers), func(i, j int) {
		potentialObservers[i], potentialObservers[j] = potentialObservers[j], potentialObservers[i]
	})

	observerId := potentialObservers[0]
	items, ok := observerMap[observerId]

	if !ok || len(items) <= 0 {
		return inatapi.Observation{}, fmt.Errorf(
			"could not find unseen observations for observer %d",
			observerId,
		)
	}

	rand.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})

	observation := items[0]

	return observation, nil

}

func (s *service) FindUnseenObservation() (inatapi.Observation, error) {
	observations, err := s.api.FetchRecentProjectObservations(s.projectID, s.pageSize, 200)

	if len(observations) <= 0 {
		if err != nil {
			return inatapi.Observation{}, fmt.Errorf("error fetching observations: %w", err)
		}

		return inatapi.Observation{}, errors.New("no unseen observations found")
	}

	o, err := s.selectUnseenObservation(observations)

	if err != nil {
		return inatapi.Observation{}, fmt.Errorf("error fetching unseen observation: %w", err)
	}

	return o, nil
}

func (s *service) MarkObservationAsSeen(
	ctx context.Context,
	o inatapi.Observation,
) (store.SeenObservation, error) {
	if !slices.Contains(s.displayedObservers, o.UserID) {
		s.displayedObservers = append(s.displayedObservers, o.UserID)
	}

	seen, err := s.store.CreateSeenObservation(ctx, o.ID)

	if err != nil {
		return store.SeenObservation{}, fmt.Errorf("error saving seen observation: %w", err)
	}

	return seen, nil
}
