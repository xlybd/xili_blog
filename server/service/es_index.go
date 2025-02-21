package service

import (
	"context"
	"server/global"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

type EsService struct{}

func (EsService *EsService) IndexCreate(indexName string, mapping *types.TypeMapping) error {
	_, err := global.ESClient.Indices.Create(indexName).Mappings(mapping).Do(context.TODO())
	return err
}

func (EsService *EsService) IndexDelete(indexName string) error {
	_, err := global.ESClient.Indices.Delete(indexName).Do(context.TODO())
	return err
}

func (EsService *EsService) IndexExists(indexName string) (bool, error) {
	return global.ESClient.Indices.Exists(indexName).Do(context.TODO())
}
