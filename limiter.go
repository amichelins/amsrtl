package amsrtl

import (
    "encoding/json"
    "fmt"
    "net"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"

    uniqueid "github.com/albinj12/unique-id"
)

// Utilizado para receber os tokens via env
type Tokens struct {
    Token string `json:"token"`
    Limit int64  `json:"limit"`
}

/*
type Tokens struct {
    Token string `json:"token"`
    Limit int64  `json:"limit"`
}

func main() {
    var Tokens []Tokens
    var dados map[string]int64 = make(map[string]int64)
    //t := `[{"wewewewe":10},{"cdflsjlsjfljlj":45}]`
    t := `[{"token":"wewewewe", "limit": 23},{"token":"cdflsjlsjfljlj", "limit": 45}]`

    err := json.Unmarshal([]byte(t), &Tokens)

    if err != nil {
        panic(err)
    }

    for _, value := range Tokens {
        fmt.Printf("%s = %d \n", value.Token, value.Limit)
        dados[value.Token] = value.Limit
    }

    for key, value := range dados {
        fmt.Printf("%s = %d \n", key, value)
    }
}


*/
const (
    // Numero maximo de requisições por IP
    LIMITER_MAX = "LIMITER_MAX"

    LIMITER_BLOCK_DURATION = "LIMITER_BLOCK_DURATION"

    // Json com o seguinte formato: ["{token1}": {valor int64},""{nome token2}\"": {valor int64}]
    // EX:
    //     ["\"{nome token1}\"": {valor int64}, ,"\"{nome token2}\"": {valor int64}]

    LIMITER_TOKENS = "LIMITER_TOKENS"
)

// Limiter Struct que da base ao Limiter
//
// storage Storage Usa a interface Storage para permitir diferentes tipos de repositorios
//
// block bool Indica se esta bloqueado ou nao, se utiliza junto com blockAt para fazer desbloqueio auto
//
// blockAt time.Time Data Hora de inicio do bloqueio
//
// blockDuration int64 Tempo em Segundos que durara o bloqueio
//
// maxRequest int64 Numero maximo de requisições quando o bloqueio é por IP
//
// tokens []map[string]int64 Array que contem o maximo de requisições por Token
//
type Limiter struct {
    storage       Storage
    block         bool
    blockAt       time.Time
    blockDuration time.Duration
    maxRequest    int64
    tokens        map[string]int64
}

// Usamos funções anonimas pela possbilidades que nos dão para configurar o Limiter
type ConF = func(*Limiter)

// NewLimiter Criamos o Limiter com maximo de requisições e com tempo de bloqueio
//            Para criar a partir de variaveis de ambiente utilizar NewEnvLimiter
// PARAMETERES
//
//     storage Storage Aqui passamos qual storage utilizaremos. Tem que implementar a interface storage
//
//     maxRequest int64 Numero maximo de requisições por segundo
//
//     blockDuration time.Duration Tempo em segundos que durara o bloqueio
//
//  RETURN
//
//     *Limiter O limiter inicializado
//
func NewLimiter(storage Storage, maxRequest int64, blockDuration time.Duration) *Limiter {
    return &Limiter{storage: storage, maxRequest: maxRequest, blockDuration: blockDuration, tokens: make(map[string]int64),
        block: false}
}

// NewEnvLimiter Criamos o Limiter baseado nos dados obtidos das variaveis de ambiente
// PARAMETERES
//
//     storage Storage Aqui passamos qual storage utilizaremos. Tem que implementar a interface storage
//
//  RETURN
//
//     *Limiter O limiter inicializado
//
func NewEnvLimiter(storage Storage) *Limiter {
    var maxRequest int64
    var blockDuration int64
    var Err error
    var sLimiter_Max string
    var sLimiter_Blk_Dur string
    var sLimiter_Token string
    var Tokens []Tokens

    // Numero maximo de requisições por IP
    sLimiter_Max = os.Getenv(LIMITER_MAX)

    // Tempo em segundos que durara o bloqueio, seja ele por ip ou por token
    sLimiter_Blk_Dur = os.Getenv(LIMITER_BLOCK_DURATION)

    // Token pelo qual sera limitado
    sLimiter_Token = os.Getenv(LIMITER_TOKENS)

    // Se a erro, utilizamos 300 segundos ( 5 minutos ) como a duração do bloqueio
    if blockDuration, Err = strconv.ParseInt(sLimiter_Blk_Dur, 0, 64); Err != nil {
        blockDuration = 300
    }

    // Se a erro, utilizamos 300 segundos ( 5 minutos ) como a duração do bloqueio
    if maxRequest, Err = strconv.ParseInt(sLimiter_Max, 0, 64); Err != nil {
        maxRequest = 100
    }

    // Criamos o limiter
    limiter := Limiter{storage: storage, maxRequest: maxRequest, blockDuration: time.Duration(blockDuration * int64(time.Second)), tokens: make(map[string]int64),
        block: false}

    // Verificamos se a dados em LIMITER_TOKENS
    // Se a e temos um json valido adidionamos os tokens a lista
    if sLimiter_Token != "" {

        err := json.Unmarshal([]byte(sLimiter_Token), &Tokens)

        if err == nil {
            for _, value := range Tokens {
                limiter.LimiterSetToken(value.Token, value.Limit)
            }
        }
    }
    return &limiter
}

// Adiciona um token com seu respetivo limitador
func (l *Limiter) LimiterSetToken(token string, limiter int64) {
    l.tokens[token] = limiter
}

// BlockAt Marca quando o bloqueio se inicio
func (l *Limiter) BlockAt() {
    l.blockAt = time.Now()
    l.block = true
}

// IsBlock Verifica se esta bloqueado e se o tempo de bloqueio esta dentro do configurado.
//         se estiver bloqueado e o tempo debloqueio espirou, ele desbloqueia
func (l *Limiter) IsBlock() bool {
    // Verificamos se Now é menor que a data hora de incio do bloqueio mais o tempo do bloqueio mais 1
    // Tudo espresado em segundos
    if l.block && time.Now().Before(l.blockAt.Add((1+l.blockDuration)*time.Second)) {
        return true
    } else {
        l.block = false
    }
    return false
}

func (l *Limiter) GetHederLimit(r *http.Request) (string, int64) {
    if len(l.tokens) > 0 {
        for Token, Limit := range l.tokens {
            if tokenValor := r.Header.Get(Token); tokenValor != "" {
                return tokenValor, Limit
            }
        }
    }
    return "", 0
}

// Run Executa o limiter sobre os dados da requisição.
func (l *Limiter) Run(w http.ResponseWriter, r *http.Request) error {
    var sChave string = getUserIP(r)
    var nLimit int64 = l.maxRequest

    // Se esta bloqueado respondo com bloqueio
    if l.IsBlock() {
        w.WriteHeader(http.StatusTooManyRequests)
        return fmt.Errorf("you have reached the maximum number of requests or actions allowed within a certain time frame")
    }

    // Context utilizado no Storage
    ctx := r.Context()

    // Verificamos se temos token para limitar
    token, limit := l.GetHederLimit(r)

    // Se o limit do token é maior que o do ip usamos os dados do token por que este tem prioridade
    if limit > l.maxRequest {
        sChave = token
        nLimit = limit
    }

    Agora := time.Now()
    Min := float64(Agora.Add(-time.Second).Unix())
    Max := float64(Agora.Unix())

    Data, err := l.storage.GetData(ctx, sChave, Min, Max)

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return fmt.Errorf("Unable to get data from storage, the error was: %s", err)
    }

    // Se data é igual o maior que o limite
    if nLimit <= Data {
        l.BlockAt() // Realiza o bloqueio e guarda quando iniciou
        w.WriteHeader(http.StatusTooManyRequests)
        return fmt.Errorf("you have reached the maximum number of requests or actions allowed within a certain time frame")
    }
    uni, _ := uniqueid.Generateid("n", 32, "Zid")

    err = l.storage.SetData(ctx, sChave, uni, Max, Min)

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return fmt.Errorf("Unable to set data in the storage, the error was: %s", err)
    }
    return nil
}

// Get the IP address of the server's connected user.
func getUserIP(r *http.Request) string {
    var userIP string
    if len(r.Header.Get("CF-Connecting-IP")) > 1 {
        return r.Header.Get("CF-Connecting-IP")

    } else if len(r.Header.Get("X-Forwarded-For")) > 1 {
        userIP = r.Header.Get("X-Forwarded-For")

    } else if len(r.Header.Get("X-Real-IP")) > 1 {
        userIP = r.Header.Get("X-Real-IP")

    } else {
        userIP = r.RemoteAddr
        if strings.Contains(userIP, ":") {
            return net.ParseIP(strings.Split(userIP, ":")[0]).String()
        } else {
            return net.ParseIP(userIP).String()
        }
    }
    return userIP
}
