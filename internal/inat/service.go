package inat

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"

	"adamolsen.dev/buggins/internal/db"
)

func getKeysFromObserverMap(m map[int64][]observation) []int64 {
	keys := make([]int64, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

type ServiceConfig struct {
	ProjectID string
	DB        *db.Queries
	PageSize  int
}

type service struct {
	ServiceConfig
	displayedObservers []int64
}

func NewService(config ServiceConfig) service {
	return service{ServiceConfig: config}
}

func (s *service) selectUnseenObservation(observations []observation) (observation, error) {
	var (
		observationIds     []int64
		unseen             []observation
		seenIds            []int64
		observerMap        map[int64][]observation = make(map[int64][]observation)
		potentialObservers []int64
	)

	for _, o := range observations {
		observationIds = append(observationIds, o.ID)
	}

	seen, err := s.DB.FindObservationsByIds(context.Background(), observationIds)

	if err != nil {
		return observation{}, fmt.Errorf("error selecting seen observations: %w", err)
	}

	for _, o := range seen {
		seenIds = append(seenIds, o.ID)
	}

	for _, o := range observations {
		if !slices.Contains(seenIds, o.ID) {
			unseen = append(unseen, o)

			items, ok := observerMap[o.UserID]

			if !ok {
				items = make([]observation, 0)
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
		return observation{}, errors.New("no unseen observations found")
	}

	if len(potentialObservers) <= 0 {
		potentialObservers = getKeysFromObserverMap(observerMap)
		s.displayedObservers = s.displayedObservers[:0]
	}

	rand.Shuffle(len(potentialObservers), func(i, j int) {
		potentialObservers[i], potentialObservers[j] = potentialObservers[j], potentialObservers[i]
	})

	observerId := potentialObservers[0]
	items, ok := observerMap[observerId]

	if !ok || len(items) <= 0 {
		return observation{}, fmt.Errorf(
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

func (s *service) FindUnseenObservation() (observation, error) {
	observations, err := fetchRecentProjectObservations(s.ProjectID, s.PageSize, 200)

	if len(observations) <= 0 {
		if err != nil {
			return observation{}, fmt.Errorf("error fetching observations: %w", err)
		}

		return observation{}, errors.New("no unseen observations found")
	}

	o, err := s.selectUnseenObservation(observations)

	if err != nil {
		return observation{}, fmt.Errorf("error fetching unseen observation: %w", err)
	}

	return o, nil
}

func (s *service) MarkObservationAsSeen(
	ctx context.Context,
	o observation,
) (db.SeenObservation, error) {
	if !slices.Contains(s.displayedObservers, o.UserID) {
		s.displayedObservers = append(s.displayedObservers, o.UserID)
	}

	seen, err := s.DB.CreateSeenObservation(ctx, o.ID)

	if err != nil {
		return db.SeenObservation{}, fmt.Errorf("error saving seen observation: %w", err)
	}

	return seen, nil
}
