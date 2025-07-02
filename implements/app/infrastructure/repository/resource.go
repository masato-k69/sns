package repository

import (
	"app/gen/pubsub"
	isearchengine "app/infrastructure/adapter/datastore/searchengine"
	"app/infrastructure/adapter/mq"
	"bytes"
	"context"
	"encoding/json"

	dfactory "app/domain/factory"
	dmodel "app/domain/model"
	drepository "app/domain/repository"
	imodel "app/infrastructure/model"

	isearch "github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	itypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/samber/lo"
)

type resourceSearchIndexRepository struct {
	resourceSearchIndexStoreConnectionRDB isearchengine.ResourceSearchIndexStoreConnection
	resourceSearchIndexStoreConnectionMQ  mq.ResourceSearchIndexStoreConnection
}

// Delete implements repository.ResourceSearchIndexRepository.
func (r *resourceSearchIndexRepository) Delete(c context.Context, id uuid.UUID) error {
	return r.resourceSearchIndexStoreConnectionMQ.Publish(c, mq.ExchangeResource, mq.RoutingKeyResourceDelete, pubsub.ResourceSearchIndex{
		ResourceId: &pubsub.UUID{Value: id.String()},
	})
}

// Update implements repository.ResourceSearchIndexRepository.
func (r *resourceSearchIndexRepository) Update(c context.Context, index dmodel.ResourceSearchIndex) error {
	return r.resourceSearchIndexStoreConnectionMQ.Publish(c, mq.ExchangeResource, mq.RoutingKeyResourceUpdate, pubsub.ResourceSearchIndex{
		ResourceId:   &pubsub.UUID{Value: index.ResourceID.String()},
		ResourceType: &pubsub.Resource{Value: index.Type.String()},
		Keyword:      index.Keyword.String(),
	})
}

// Create implements repository.ResourceSearchIndexRepository.
func (r *resourceSearchIndexRepository) Create(c context.Context, index dmodel.ResourceSearchIndex) error {
	return r.resourceSearchIndexStoreConnectionMQ.Publish(c, mq.ExchangeResource, mq.RoutingKeyResourceCreate, pubsub.ResourceSearchIndex{
		ResourceId:   &pubsub.UUID{Value: index.ResourceID.String()},
		ResourceType: &pubsub.Resource{Value: index.Type.String()},
		Keyword:      index.Keyword.String(),
	})
}

// List implements repository.ResourceSearchIndexRepository.
func (r *resourceSearchIndexRepository) List(c context.Context, resourceTypes []dmodel.Resource, freeword string, page dmodel.Range) ([]dmodel.ResourceSearchIndex, error) {
	response, err := r.resourceSearchIndexStoreConnectionRDB.Client().
		Search().
		Index(imodel.ResourceSearchIndex{}.Index()).
		Request(&isearch.Request{
			From: &page.Offset,
			Size: &page.Limit,
			Query: &itypes.Query{
				Bool: &itypes.BoolQuery{
					Must: []itypes.Query{
						{
							Terms: &itypes.TermsQuery{
								TermsQuery: map[string]itypes.TermsQueryField{
									"type": lo.Map(resourceTypes, func(resourceType dmodel.Resource, _ int) string { return resourceType.String() }),
								},
							},
						},
						{
							Match: map[string]itypes.MatchQuery{
								"keyword": {
									Query: freeword,
								},
							},
						},
					},
				},
			},
		}).
		Do(c)

	if err != nil {
		return nil, err
	}

	dIndexes := []dmodel.ResourceSearchIndex{}
	for _, hit := range response.Hits.Hits {
		bytes, err := hit.Source_.MarshalJSON()

		if err != nil {
			return nil, err
		}

		index := imodel.ResourceSearchIndex{}
		if err := json.Unmarshal(bytes, &index); err != nil {
			return nil, err
		}

		dIndex, err := dfactory.NewResourceSearchIndex(index.ResourceID, index.Type, index.Keyword)
		if err != nil {
			return nil, err
		}

		dIndexes = append(dIndexes, *dIndex)
	}

	return dIndexes, nil
}

func NewResourceSearchIndexRepository(i *do.Injector) (drepository.ResourceSearchIndexRepository, error) {
	resourceSearchIndexStoreConnection := do.MustInvoke[isearchengine.ResourceSearchIndexStoreConnection](i)
	resourceSearchIndexStoreConnectionMQ := do.MustInvoke[mq.ResourceSearchIndexStoreConnection](i)
	return &resourceSearchIndexRepository{
		resourceSearchIndexStoreConnectionRDB: resourceSearchIndexStoreConnection,
		resourceSearchIndexStoreConnectionMQ:  resourceSearchIndexStoreConnectionMQ,
	}, nil
}

type resourceSearchIndexRepositoryForAsync struct {
	resourceSearchIndexStoreConnection isearchengine.ResourceSearchIndexStoreConnection
}

// Create implements repository.ResourceSearchIndexRepository.
func (r *resourceSearchIndexRepositoryForAsync) Create(c context.Context, index dmodel.ResourceSearchIndex) error {
	return r.save(c, index)
}

// Delete implements repository.ResourceSearchIndexRepository.
func (r *resourceSearchIndexRepositoryForAsync) Delete(c context.Context, id uuid.UUID) error {
	if _, err := r.resourceSearchIndexStoreConnection.Client().
		Delete(imodel.ResourceSearchIndex{}.Index(), id.String()).
		Do(c); err != nil {
		return err
	}

	return nil
}

// List implements repository.ResourceSearchIndexRepository.
func (r *resourceSearchIndexRepositoryForAsync) List(c context.Context, resourceTypes []dmodel.Resource, freeword string, page dmodel.Range) ([]dmodel.ResourceSearchIndex, error) {
	panic("unimplemented")
}

// Update implements repository.ResourceSearchIndexRepository.
func (r *resourceSearchIndexRepositoryForAsync) Update(c context.Context, index dmodel.ResourceSearchIndex) error {
	return r.save(c, index)
}

func (r *resourceSearchIndexRepositoryForAsync) save(c context.Context, index dmodel.ResourceSearchIndex) error {
	iIndex := imodel.ResourceSearchIndex{
		ResourceID: index.ResourceID.String(),
		Type:       index.Type.String(),
		Keyword:    index.Keyword.String(),
	}

	bin, err := json.Marshal(iIndex)
	if err != nil {
		return err
	}

	if _, err := r.resourceSearchIndexStoreConnection.Client().
		Index(iIndex.Index()).
		Id(iIndex.ResourceID).
		Raw(bytes.NewReader(bin)).
		Do(c); err != nil {
		return err
	}

	return nil
}

func NewResourceSearchIndexRepositoryForAsync(i *do.Injector) (drepository.ResourceSearchIndexRepository, error) {
	resourceSearchIndexStoreConnection := do.MustInvoke[isearchengine.ResourceSearchIndexStoreConnection](i)
	return &resourceSearchIndexRepositoryForAsync{
		resourceSearchIndexStoreConnection: resourceSearchIndexStoreConnection,
	}, nil
}
