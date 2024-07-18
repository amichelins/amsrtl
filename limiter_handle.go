package amsrtl

import (
    "encoding/json"
    "net/http"
)

// Handle Recebe o limiter e o http.Handler que esta sendo executado.
//        chama ao Run para executar a logica de limitação se tiver bloqueio retorna,
//        se não executa o handler enviado
func Handle(limiter *Limiter, handle http.Handler) http.Handler {
    return http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
            Err := limiter.Run(w, r)

            if Err != nil {
                //fmt.Printf("%s", Err)
                json.NewEncoder(w).Encode(struct {
                    Error string `json:"error"`
                }{
                    Error: Err.Error(),
                })
                return
            }

            handle.ServeHTTP(w, r)
        },
    )

}
