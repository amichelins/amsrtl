package amsrtl

import (
    "context"
)

// Storage, Interface que todos os storage tem que cumprir, seja eles
// Redis, ou um outro tipo de alamacenamento
type Storage interface {
    // context.Context Contexto da chamada
    // string Chave do dado a ser lido
    // time.Duration Período de tempo do rate
    // RETURN
    //     int64 Contagem das interações
    //     error Erro quando ouver ou nil
    GetData(context.Context, string, float64, float64) (int64, error)

    // context.Context Contexto da chamada

    // RETURN
    //     error Erro quando ouver ou nil
    SetData(context.Context, string, string, float64, float64) error
}
