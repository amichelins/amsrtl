# amsrtl
Rate Limiter

Para Inicializar com variáveis de ambiente.
Elas são :

    LIMITER_MAX Limite maximo de conexões por IP
    LIMITER_BLOCK_DURATION Duração em segundos do bloqueio
    LIMITER_TOKENS Lista de token validos em formato JSON.
       EX: [{"token":"TOKENA","limit": 10},{"token":"TOKENB","limit": 10}]

Usar:
    {storage} é um interface que permite varios tipos de storage
    amsrtl.NewEnvLimiter({storage})