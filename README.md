# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

### Оптимизация

После оптимизации кода относительно базового профиля (pprof -top -diff_base=profiles/base.pprof profiles/result.pprof):

```
File: integrationbench.test
Type: alloc_space
Time: 2026-05-01 20:19:21 MSK
Showing nodes accounting for 585.22MB, 4.76% of 12293.42MB total
Dropped 50 nodes (cum <= 61.47MB)
      flat  flat%   sum%        cum   cum%
  299.14MB  2.43%  2.43%   299.14MB  2.43%  bufio.NewReaderSize (inline)
  -63.01MB  0.51%  1.92%   -80.51MB  0.65%  net/url.(*URL).JoinPath
   42.51MB  0.35%  2.27%    42.51MB  0.35%  strings.(*Builder).grow
      34MB  0.28%  2.54%       34MB  0.28%  net/http/httptest.NewRecorder (inline)
   32.01MB  0.26%  2.80%    32.01MB  0.26%  net/http.(*Request).SetPathValue (inline)
   29.01MB  0.24%  3.04%    29.01MB  0.24%  encoding/json.(*Decoder).refill
     -27MB  0.22%  2.82%      -27MB  0.22%  net/url.parse
   25.01MB   0.2%  3.02%    25.01MB   0.2%  github.com/golang-jwt/jwt/v5.NewWithClaims (inline)
   24.51MB   0.2%  3.22%    24.51MB   0.2%  net/http.Header.Clone (inline)
   21.51MB  0.17%  3.40%    21.51MB  0.17%  net/http.(*Request).WithContext (partial-inline)
   21.51MB  0.17%  3.57%    21.51MB  0.17%  net/textproto.MIMEHeader.Add (inline)
   18.50MB  0.15%  3.72%    18.50MB  0.15%  encoding/base64.(*Encoding).EncodeToString
   18.01MB  0.15%  3.87%    26.51MB  0.22%  net/http.readRequest
   17.50MB  0.14%  4.01%    17.50MB  0.14%  net/textproto.MIMEHeader.Set (inline)
   16.50MB  0.13%  4.15%    16.50MB  0.13%  crypto/internal/fips140/sha256.New (inline)
      15MB  0.12%  4.27%       15MB  0.12%  github.com/google/uuid.UUID.String (inline)
      12MB 0.098%  4.37%    28.50MB  0.23%  crypto/internal/fips140/hmac.New[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }]
   11.50MB 0.094%  4.46%    11.50MB 0.094%  github.com/golang-jwt/jwt/v5.NewNumericDate (inline)
      11MB 0.089%  4.55%   112.02MB  0.91%  github.com/Irongoshan-ux/url-shortener/internal/auth.makeCookie
       9MB 0.073%  4.62%        9MB 0.073%  encoding/json.NewDecoder (inline)
       9MB 0.073%  4.70%    11.50MB 0.094%  bytes.(*Buffer).grow
    8.50MB 0.069%  4.76%    59.01MB  0.48%  github.com/golang-jwt/jwt/v5.(*Token).SignedString
    7.50MB 0.061%  4.83%     7.50MB 0.061%  encoding/json.(*decodeState).literalStore
      -7MB 0.057%  4.77%       -6MB 0.049%  encoding/json.Marshal
      -7MB 0.057%  4.71%   -17.50MB  0.14%  path.Join
    6.50MB 0.053%  4.76%    38.50MB  0.31%  github.com/golang-jwt/jwt/v5.(*SigningMethodHMAC).Sign
   -6.50MB 0.053%  4.71%       -5MB 0.041%  encoding/json.mapEncoder.encode
    5.50MB 0.045%  4.76%     5.50MB 0.045%  github.com/Irongoshan-ux/url-shortener/internal/auth.secretBytes (inline)
    5.50MB 0.045%  4.80%        6MB 0.049%  github.com/golang-jwt/jwt/v5.NumericDate.MarshalJSON
   -5.50MB 0.045%  4.76%    -5.50MB 0.045%  path.(*lazybuf).append (inline)
      -5MB 0.041%  4.72%       -5MB 0.041%  context.WithValue
      -5MB 0.041%  4.68%       -5MB 0.041%  path.(*lazybuf).string (inline)
    4.50MB 0.037%  4.71%   584.40MB  4.75%  github.com/Irongoshan-ux/url-shortener/internal/integrationbench.BenchmarkWorkloadMixed
      -4MB 0.033%  4.68%   -71.34MB  0.58%  github.com/Irongoshan-ux/url-shortener/internal/handler.(*Handler).ShortenURLJSON
    3.50MB 0.028%  4.71%    40.52MB  0.33%  net/http.Redirect
    3.50MB 0.028%  4.74%     3.50MB 0.028%  crypto/internal/fips140/sha256.(*Digest).Sum
       3MB 0.024%  4.76%   206.72MB  1.68%  github.com/Irongoshan-ux/url-shortener/internal/integrationbench.BenchmarkWorkloadMixed.CookieMiddleware.func1.1
       3MB 0.024%  4.78%        3MB 0.024%  github.com/go-chi/chi/v5.(*Mux).nextRoutePath (inline)
      -3MB 0.024%  4.76%       -3MB 0.024%  github.com/google/uuid.NewRandomFromReader
   -0.50MB 0.0041%  4.76%        7MB 0.057%  encoding/json.(*decodeState).object
    0.50MB 0.0041%  4.76%   325.15MB  2.64%  net/http/httptest.NewRequestWithContext
         0     0%  4.76%   299.14MB  2.43%  bufio.NewReader (inline)
         0     0%  4.76%       11MB 0.089%  bytes.(*Buffer).Write
         0     0%  4.76%    16.50MB  0.13%  crypto.Hash.New
         0     0%  4.76%    28.50MB  0.23%  crypto/hmac.New
         0     0%  4.76%    16.50MB  0.13%  crypto/hmac.New.UnwrapNew[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }].func1
         0     0%  4.76%     3.50MB 0.028%  crypto/internal/fips140/hmac.(*HMAC).Sum
         0     0%  4.76%    16.50MB  0.13%  crypto/sha256.New
         0     0%  4.76%    35.01MB  0.28%  encoding/json.(*Decoder).Decode
         0     0%  4.76%    28.01MB  0.23%  encoding/json.(*Decoder).readValue
         0     0%  4.76%     8.50MB 0.069%  encoding/json.(*Encoder).Encode
         0     0%  4.76%        7MB 0.057%  encoding/json.(*decodeState).unmarshal
         0     0%  4.76%        7MB 0.057%  encoding/json.(*decodeState).value
         0     0%  4.76%     5.50MB 0.045%  encoding/json.marshalerEncoder
         0     0%  4.76%     5.50MB 0.045%  encoding/json.structEncoder.encode
         0     0%  4.76%        4MB 0.033%  fmt.Fprintln
         0     0%  4.76%       12MB 0.098%  github.com/Irongoshan-ux/url-shortener/internal/auth.GenerateUserID
         0     0%  4.76%   124.02MB  1.01%  github.com/Irongoshan-ux/url-shortener/internal/auth.GetOrCreateUserID
         0     0%  4.76%    40.52MB  0.33%  github.com/Irongoshan-ux/url-shortener/internal/handler.(*Handler).Redirect
         0     0%  4.76%  -133.02MB  1.08%  github.com/Irongoshan-ux/url-shortener/internal/handler.(*Handler).buildFullURL
         0     0%  4.76%    14.50MB  0.12%  github.com/Irongoshan-ux/url-shortener/internal/validation.ParseHTTPURL
         0     0%  4.76%   -27.82MB  0.23%  github.com/go-chi/chi/v5.(*Mux).Mount.func1
         0     0%  4.76%   207.73MB  1.69%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0%  4.76%     4.19MB 0.034%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0%  4.76%    32.01MB  0.26%  github.com/go-chi/chi/v5.setPathValue
         0     0%  4.76%       25MB   0.2%  github.com/golang-jwt/jwt/v5.(*Token).EncodeSegment (inline)
         0     0%  4.76%       -3MB 0.024%  github.com/google/uuid.New (inline)
         0     0%  4.76%       -3MB 0.024%  github.com/google/uuid.NewRandom
         0     0%  4.76%    43.51MB  0.35%  net/http.(*Cookie).String
         0     0%  4.76%   206.72MB  1.68%  net/http.HandlerFunc.ServeHTTP
         0     0%  4.76%    21.51MB  0.17%  net/http.Header.Add (inline)
         0     0%  4.76%    17.50MB  0.14%  net/http.Header.Set (inline)
         0     0%  4.76%    26.51MB  0.22%  net/http.ReadRequest
         0     0%  4.76%    65.02MB  0.53%  net/http.SetCookie
         0     0%  4.76%       11MB 0.089%  net/http/httptest.(*ResponseRecorder).Write
         0     0%  4.76%    24.51MB   0.2%  net/http/httptest.(*ResponseRecorder).WriteHeader
         0     0%  4.76%   325.15MB  2.64%  net/http/httptest.NewRequest (inline)
         0     0%  4.76%      -18MB  0.15%  net/url.(*URL).String
         0     0%  4.76%  -150.02MB  1.22%  net/url.JoinPath
         0     0%  4.76%   -38.01MB  0.31%  net/url.Parse
         0     0%  4.76%       11MB 0.089%  net/url.ParseRequestURI
         0     0%  4.76%   -10.50MB 0.085%  path.Clean
         0     0%  4.76%        5MB 0.041%  reflect.(*MapIter).Key
         0     0%  4.76%       -4MB 0.033%  reflect.(*MapIter).Value
         0     0%  4.76%    42.51MB  0.35%  strings.(*Builder).Grow
         0     0%  4.76%   584.89MB  4.76%  testing.(*B).launch
         0     0%  4.76%   584.40MB  4.75%  testing.(*B).runN
```

**Разбор результатов (alloc_space, diff к базовому профилю).** Сводка по таблице выше: основной выигрыш по аллокациям связан с горячими путями обработчика и сборки URL - в частности, `net/url.JoinPath` и связанные `Parse`/`String` дают суммарное снижение порядка сотен мегабайт в diff, что соответствует уходу от дорогих операций склейки пути в пользу более дешёвой схемы формирования полного short URL (`buildFullURL`). По `ShortenURLJSON` виден отрицательный вклад в аллокации — меньше временных объектов на запрос сокращения. При этом в топе по-прежнему остаются инфраструктурные узлы бенчмаркинга: `httptest.NewRequest`, буферы `bufio`, клонирование заголовков, JWT/cookie middleware — они отражают не только прикладной код, но и накладные расходы тестового HTTP-стека. Итог: оптимизация позволила уменьшить аллокации при построении URL и при обработке запроса на сокращение.
