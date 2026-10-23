# Календарь звонков (продолжение)


[![hexlet-check](https://github.com/ekup/ai-for-developers-project-387/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/ekup/ai-for-developers-project-387/actions)

**Приложение**: https://ai-for-developers-project-387-krp5.onrender.com — демо-прод на Render, данные не переживают рестарт (см. раздел [«Прод»](#прод)).

Интегрируйте работу агентов в GitHub проект

Учебный проект Хекслета: https://ru.hexlet.io/programs/ai-for-developers
Как это должно работать: https://files.hexlet.app/a/2ipc5m

## Стек

**Бэкенд**

- Go 1.27
- [chi](https://github.com/go-chi/chi) — HTTP-роутер
- SQLite ([modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite), чистый Go, без cgo)
- Линтеры: `go vet`, `staticcheck`
- Тесты: `go test`

**Фронтенд**

- TypeScript, [Vite](https://vite.dev/), React
- [Mantine](https://mantine.dev/) — UI-библиотека
- Линтеры: ESLint (`eslint-plugin-react`, `eslint-plugin-react-hooks`), Prettier
- Тесты: [Vitest](https://vitest.dev/) + Testing Library

## Структура

```
.
├── backend/            # Go-сервис (chi + sqlite), отдаёт API и статику фронтенда
│   ├── cmd/server/     # точка входа
│   └── internal/       # config, server, store, api (сгенерированные артефакты)
├── frontend/           # Vite + React + Mantine
│   └── src/            # лендинг, тесты и src/api (клиентский SDK)
├── api/                # контракт API на TypeSpec → OpenAPI (api/openapi/openapi.yaml)
├── scripts/            # дымовые тесты: smoke.sh и docker-smoke.sh
├── Dockerfile          # мультистейдж: фронтенд + бэкенд в одном образе
├── render.yaml         # прод-конфигурация Render (блюпринт)
└── Makefile            # команды запуска, тестов и линтеров
```

## Установка

```bash
git clone https://github.com/ekup/ai-for-developers-project-387.git
cd ai-for-developers-project-387

# зависимости фронтенда и бэкенда
make install

# линтер для Go
go install honnef.co/go/tools/cmd/staticcheck@latest
```

## Запуск локально

Бэкенд и фронтенд запускаются раздельно. Бэкенд слушает `:8080`, Vite — `:5173` и
проксирует `/api` на бэкенд.

```bash
# терминал 1 — бэкенд
make dev-backend

# терминал 2 — фронтенд (http://localhost:5173)
make dev-frontend
```

Проверка бэкенда:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

## Тесты и линтеры

```bash
make test           # тесты бэкенда и фронтенда
make lint           # go vet + staticcheck, ESLint + Prettier

make backend-test   # только тесты бэкенда
make backend-lint   # только линтеры бэкенда
make frontend-test  # только тесты фронтенда
make frontend-lint  # только линтеры фронтенда

make smoke          # дымовой тест: приложение поднимается и отвечает
```

## API-контракт

Контракт API описан на [TypeSpec](https://typespec.io/) в `api/` и служит единственным источником истины для запросов и ответов. Из него генерируются OpenAPI-спецификация, клиентский SDK фронтенда и серверные артефакты Go.

```bash
make api-gen        # TypeSpec → api/openapi/openapi.yaml → SDK + Go-артефакты
make api-check      # то же самое + проверка, что артефакты не разошлись с контрактом
make api-docs       # открыть интерактивную документацию API (Swagger UI) в браузере
```

Что получается на выходе:

- `api/openapi/openapi.yaml` — OpenAPI-спецификация (эмиттер `@typespec/openapi3`);
- `frontend/src/api/generated/` — типизированный SDK (`@hey-api/openapi-ts`), поверх него — обёртка `frontend/src/api/client.ts`;
- `backend/internal/api/api.gen.go` — модели, валидация запросов и интерфейс маршрутов chi (`oapi-codegen`).

Сгенерированные файлы коммитятся, поэтому сборка не требует запуска генерации. Если правишь контракт — запусти `make api-gen` и закоммить результат: CI-задача `api` падает, если артефакты разошлись с `api/*.tsp`.

## Docker

Приложение собирается в один образ: бэкенд раздаёт собранную статику фронтенда.

```bash
make docker-build
make docker-run     # http://localhost:8080
make docker-smoke   # собрать образ, поднять контейнер и проверить /health
```

Или напрямую:

```bash
docker build -t call-booking:latest .
docker run --rm -p 8080:8080 -v call-booking-data:/app/data call-booking:latest
```

### Переменные окружения

| Переменная      | По умолчанию  | Назначение                          |
| --------------- | ------------- | ----------------------------------- |
| `HOST`          | `0.0.0.0`     | адрес HTTP-сервера                  |
| `PORT`          | `8080`        | порт HTTP-сервера                   |
| `DATABASE_PATH` | `data/app.db` | путь к файлу SQLite                 |
| `STATIC_DIR`    | `static`      | каталог со собранной статикой фронта |
| `TZ`            | пояс хоста    | пояс записи: рабочие часы и окно записи считаются в нём. В проде и dev — `Europe/Moscow` (см. `docs/adr/0003`) |

## Прод

Прод живёт на [Render](https://render.com/): один web-сервис `ai-for-developers-project-387` в регионе Frankfurt, который собирается из корневого `Dockerfile` и раздаёт API и статику фронтенда. Пуш в `main` запускает деплой только после зелёных проверок ветки, а Render считает сервис живым, пока `/health` отвечает `200`.

- приложение: https://ai-for-developers-project-387-krp5.onrender.com
- здоровье: https://ai-for-developers-project-387-krp5.onrender.com/health

Пояс записи в проде задан переменной `TZ=Europe/Moscow` (см. `docs/adr/0003`): в runtime-образ добавлен `tzdata`, поэтому рабочие часы и окно записи считаются по Москве, а не по UTC.

Compute-план — `free`, поэтому это **демо-прод**:

- файловая система эфемерная: база `app.db` стирается при каждом редеплое, рестарте и спин-дауне. Созданные типы событий и записи гостей не переживают даже 15 минут без трафика; схема и seed-типы пересоздаются при старте;
- после спин-дауна первый запрос ждёт около минуты, пока сервис поднимается.

Авторизации в сервисе нет (см. `docs/spec.md`, раздел 9), поэтому владельческие страницы закрыты от индексации в `frontend/public/robots.txt`, но остаются доступными по прямой ссылке.

`render.yaml` описывает ту же конфигурацию, что и текущий сервис, но блюпринт не «усыновляет» сервис, созданный вручную в Dashboard. Применение через **New → Blueprint** создаст новый сервис и, вероятно, другой субдомен — это способ воспроизвести прод с нуля, а не обновить текущий.

## CI

GitHub Actions (`.github/workflows/ci.yml`) на каждый push и pull request:

- бэкенд: `go vet`, `staticcheck`, `go test`;
- фронтенд: ESLint, Prettier, Vitest;
- дымовой тест: приложение собирается, поднимается и отвечает на `/health`;
- Docker: образ собирается, контейнер стартует и слушает порт из `PORT`.

## Работа агентов OpenCode

Пять воркфлоу запускают агента OpenCode. Модель у всех одна — `opencode-go/deepseek-v4.1-flash`
(провайдер `opencode-go` по подписке OpenCode Go, секрет `OPENCODE_API_KEY`). Прогоны — в разделе
[Actions](https://github.com/ekup/ai-for-developers-project-387/actions); в таблице ниже — прямые
ссылки на конкретный воркфлоу.

| Воркфлоу                                                             | Событие                                                                                        | Модель                             | Назначение                                                      | Где смотреть прогоны                                                                                                |
| -------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- | ---------------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| [`opencode.yml`](.github/workflows/opencode.yml)                     | комментарий с `/oc` или `/opencode` (`issue_comment`, `pull_request_review_comment`)           | `opencode-go/deepseek-v4.1-flash`  | правит код по задаче из комментария, коммитит и открывает PR    | [Actions → opencode](https://github.com/ekup/ai-for-developers-project-387/actions/workflows/opencode.yml)             |
| [`opencode-review.yml`](.github/workflows/opencode-review.yml)       | `pull_request`: opened, synchronize, reopened, ready_for_review                                | `opencode-go/deepseek-v4.1-flash`  | ревью диффа PR и предложения по улучшениям                      | [Actions → opencode-review](https://github.com/ekup/ai-for-developers-project-387/actions/workflows/opencode-review.yml) |
| [`opencode-triage.yml`](.github/workflows/opencode-triage.yml)       | `issues`: opened (автор старше 30 дней)                                                        | `opencode-go/deepseek-v4.1-flash`  | подсказки по issue: ссылки на доки, обработка ошибок            | [Actions → Issue Triage](https://github.com/ekup/ai-for-developers-project-387/actions/workflows/opencode-triage.yml)     |
| [`opencode-scheduled.yml`](.github/workflows/opencode-scheduled.yml) | расписание: понедельник 09:00 UTC; `workflow_dispatch`: ручной запуск                          | `opencode-go/deepseek-v4.1-flash`  | обзор `TODO`/`FIXME`/`HACK` и заведение issue с отчётом         | [Actions → Scheduled OpenCode Task](https://github.com/ekup/ai-for-developers-project-387/actions/workflows/opencode-scheduled.yml) |
| [`opencode-workflow-dispatch.yml`](.github/workflows/opencode-workflow-dispatch.yml) | `workflow_dispatch`: ручной запуск                                             | `opencode-go/deepseek-v4.1-flash` (переопределяется полем `model`) | произвольная задача из поля `prompt`            | [Actions → Manual OpenCode Task](https://github.com/ekup/ai-for-developers-project-387/actions/workflows/opencode-workflow-dispatch.yml) |

### Команды вызова и доступ

- Агента зовут комментарием, начинающимся с `/oc` или `/opencode` (либо этими же словами через пробел в середине текста). Набор задан явно через `mentions: "/oc,/opencode"` в `opencode.yml` — это дефолт экшена, он оставлен, чтобы не ломать уже написанные команды.
- Воркфлоу не срабатывает на каждый комментарий подряд: без команды в теле job пропускается (в списке прогонов это видно как `skipped`).
- События от ботов и от самого приложения OpenCode (его логины заканчиваются на `[bot]`) отсекаются условием `!endsWith(github.actor, '[bot]')` и `user.type != 'Bot'` во всех пяти воркфлоу.
- Репозиторий публичный, поэтому звать агента командой могут только `OWNER`, `MEMBER` и `COLLABORATOR` (проверка `author_association`). Ручные запуски (`workflow_dispatch`) и так доступны лишь пользователям с правом записи: в `opencode-scheduled.yml` по кнопке запускается тот же обзор `TODO`/`FIXME`/`HACK`, в `opencode-workflow-dispatch.yml` — произвольная задача из поля `prompt`. В обоих воркфлоу дополнительно проверяется, что запуск идёт не из форка.

### Права

`permissions` объявлены отдельно в каждом воркфлоу. `write` выдаётся там, где агент создаёт коммиты, задачи или PR; у `opencode-scheduled.yml` набор расширен по требованию учебного задания:

| Воркфлоу                        | `permissions`                                                              | Почему так                                                                                          |
| ------------------------------- | -------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| `opencode.yml`                  | `id-token: write`, `contents: read`, `pull-requests: write`                | коммиты и PR создаёт токен приложения OpenCode (обмен OIDC); `pull-requests: write` нужен шагу «Нормализовать заголовок PR» |
| `opencode-review.yml`           | `id-token: write`, `contents: read`, `pull-requests: write`                | `id-token` — требование учебного задания; при `use_github_token: true` OIDC фактически не используется, а комментарии в разговоре PR требуют прав именно на `pull-requests` |
| `opencode-triage.yml`           | `id-token: write`, `contents: read`                                        | комментарий создаёт токен приложения (OIDC), воркфлоу на запись ничего не просит                    |
| `opencode-scheduled.yml`        | `id-token: write`, `contents: write`, `pull-requests: write`, `issues: write` | требование учебного задания: `id-token` и write в трёх областях; текущий промпт заводит только issue |
| `opencode-workflow-dispatch.yml`| `id-token: write`, `contents: write`, `pull-requests: write`, `issues: write` | произвольная задача может дойти до коммита, PR и issue; `id-token` — требование задания             |

### Share

Репозиторий публичный, и для публичных репозиториев экшен по умолчанию расшаривает сессию агента (`share` = true). В первых PR это уже проявилось: в телах #4, #5 и #7 оставались публичные ссылки вида `opencode.ai/s/...`. Решение — **share выключен** (`share: false` во всех воркфлоу): сессии агента наружу не публикуются, чтобы не выносить промпты и содержимое репозитория. Ссылка на прогон Actions в комментарии при этом остаётся.

### Самооценка

Что получилось с первого прохода:

- Триаж issue отработал на #2 и #6 без замечаний.
- Issue #6 «Отмена бронирования» → PR #7 (`feat: отмена записи владельцем`) смержен без переработок.
- Обзор `TODO`/`FIXME`/`HACK` в ручном прогоне сразу дал issue #11 со структурированным отчётом.

Где потребовались итерации:

- Issue #3 «Не хочу созваниваться в выходные»: первый PR #4 закрыт, работу пришлось переделать — приняли только PR #5.
- `opencode-review` падал с 403 при попытке комментировать PR, пока в воркфлоу не выдали `pull-requests: write` (коммит `900b9e9`).
- «Manual OpenCode Task» один раз завис на 34 минуты на запросе разрешения (агент искал креды `gh`), пока не добавили `timeout-minutes` и проброс `GITHUB_TOKEN` (коммит `5fc73da`).
- Коммиты агента сначала не соответствовали Conventional Commits — добавили правило в `AGENTS.md` и шаг «Нормализовать заголовок PR» (коммиты `3a5813e`, `b9d76e2`).

---

<details>
<summary>Автоматические тесты Хекслета</summary>

Тесты запускаются на каждый коммит. За запуск отвечает файл `.github/workflows/hexlet-check.yml` — не удаляйте и не переименовывайте ни его, ни репозиторий.

</details>

## О Хекслете

[Хекслет](https://ru.hexlet.io/) — школа программирования: авторские программы обучения с практикой, поддержкой наставников и реальными проектами, которые остаются в резюме. Этот репозиторий — один из таких проектов.
