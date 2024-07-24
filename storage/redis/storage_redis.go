package storage_redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Implementação do Storage interface para o Redis
// redis  *redis.Client Cliente Redis para acessar o servidor
// autoLimpar bool Se true o sistema limpa os dados já utilizados, se false, terá que ser chamado o metodo Limpar
type StorageRedis struct {
    redis      *redis.Client
    autoLimpar bool
}

func NewRedisStorage(redis *redis.Client, autoLimpar bool) *StorageRedis {
    return &StorageRedis{redis: redis, autoLimpar: autoLimpar}
}

// SetData Grava os dados no ser ordenado indicado por Chave
// PARAMETERS
//     ctx context.Context Contexto da operação
//
//     Chave string Identificador do sorted SET onde gravaremos os elementos. NOTA: Pode ser um IP, token, etc
//
//     Id string, Id do item adicionado, corresponde a redis.Z.Member
//
//     Valor float64, Valor do item adicionado, corresponde a redis.Z.Score
//
//     Tempo float64, Valor utilizado para controlar a janela de limitação
//
//
func (sr *StorageRedis) SetData(ctx context.Context, Chave string, Id string, Valor float64, Tempo float64) error {
    // Usamos o ZAdd para guardar na chave os dados do rate
    // Os itens do Zadd são Structs do tipo redis.Z que tem um Score ( float64 ) e um Member string

    sr.redis.ZAdd(ctx, Chave, redis.Z{Score: Valor, Member: Id})

    // Se autoLimpar = true, limpamos os dados passados
    if sr.autoLimpar {
        sr.redis.ZRemRangeByScore(ctx, Chave, "0", fmt.Sprintf("%f", (Valor-Tempo)))
    }
    return nil
}

func (sr *StorageRedis) GetData(ctx context.Context, Chave string, Min float64, Max float64) (int64, error) {

    nCount, Err := sr.redis.ZCount(ctx, Chave, fmt.Sprintf("%f", Min), fmt.Sprintf("%f", Max)).Result()

    if Err != nil {
        return 0, Err
    }

    return nCount, nil
}
