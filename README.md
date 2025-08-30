
---

## Быстрый старт: запуск и проверка с нуля

Требования:
- Docker, Docker Compose
- Go 1.22+
- jq (для удобного вывода)

1) Клонирование
```bash
git clone https://github.com/arasvet/microtube.git
cd microtube
```

2) Поднять инфраструктуру
```bash
docker compose up -d db redis
```

3) Миграции (если не применились из init)
```bash
docker exec -i microtube-db psql -U app -d microtube < migrations/sql/0001_init.up.sql
docker exec -i microtube-db psql -U app -d microtube < migrations/sql/0002_fts_indexes.up.sql
docker exec -i microtube-db psql -U app -d microtube < migrations/sql/0003_videos_fts_trigger.up.sql
```

4) Сид данных (≥1000 видео, ≥50k событий)
```bash
export POSTGRES_USER=app POSTGRES_PASSWORD=app POSTGRES_DB=microtube POSTGRES_HOST=localhost POSTGRES_PORT=5432
go run ./cmd/seed
```

5) Запуск API
```bash
export API_HTTP_PORT=8080
export JWT_SECRET=devsecret
# пока без админов
env | grep ^ADMINS >/dev/null || export ADMINS=

go build ./cmd/api && ./api
```

### Проверка ручек (curl)

Health и OpenAPI:
```bash
curl -s http://localhost:8080/healthz
curl -s http://localhost:8080/openapi.yaml | head -20
```

Регистрация и логин:
```bash
# регистрация (409 при повторе — email уже существует)
curl -i -s -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"pass"}'

# логин
LOGIN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"pass"}')

echo "$LOGIN" | jq .
TOKEN=$(echo "$LOGIN" | jq -r .token)
echo "TOKEN: ${#TOKEN} chars"
```

Получить user_id (sub) из JWT:
```bash
USER_ID=$(python3 - <<'PY'
import os,base64,json
tok=os.environ.get('TOKEN','')
parts=tok.split('.')
payload=parts[1] + '='*((4-len(parts[1])%4)%4)
print(json.loads(base64.urlsafe_b64decode(payload))['sub'])
PY
)
echo "USER_ID=$USER_ID"
```

События (идемпотентность):
```bash
# невалидное тело -> 422
curl -i -s -X POST http://localhost:8080/events -H 'Content-Type: application/json' -d '{}'

# валидное событие
NOW=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
VID=$(docker exec -i microtube-db psql -U app -d microtube -t -c "select id from app.videos limit 1" | tr -d '[:space:]')
EID=$(uuidgen)

curl -i -s -X POST http://localhost:8080/events -H 'Content-Type: application/json' -d "{
  \"event_id\":\"$EID\",
  \"ts\":\"$NOW\",
  \"type\":\"view_start\",
  \"session_id\":\"sess-1\",
  \"video_id\":\"$VID\"
}"

# повтор тем же event_id -> 200
curl -i -s -X POST http://localhost:8080/events -H 'Content-Type: application/json' -d "{
  \"event_id\":\"$EID\",
  \"ts\":\"$NOW\",
  \"type\":\"view_start\",
  \"session_id\":\"sess-1\",
  \"video_id\":\"$VID\"
}"
```

Поиск (FTS + trigram):
```bash
curl -s "http://localhost:8080/search?q=go&limit=5" | jq '.results[0:3]'
```

Фиды:
```bash
curl -s "http://localhost:8080/videos/feed?type=popular&limit=5"   | jq '.videos[0:3]'
curl -s "http://localhost:8080/videos/feed?type=commented&limit=3" | jq '.videos[0:3]'
curl -s "http://localhost:8080/videos/feed?type=random&limit=3"    | jq '.videos[0:3]'
```

Рекомендации:
```bash
# холодные (гость)
curl -s "http://localhost:8080/recommendations?session_id=test-session&limit=5" | jq '.type,.total'
# персональные
curl -s "http://localhost:8080/recommendations?user_id=$USER_ID&limit=5" | jq '.type,.total'
```

Статистика (только админы):
```bash
# без авторизации -> 403
curl -i -s "http://localhost:8080/stats/overview?top=3" | head -5
# с токеном, но ещё не админ -> 403
curl -i -s "http://localhost:8080/stats/overview?top=3" -H "Authorization: Bearer $TOKEN" | head -5
```

Сделать пользователя админом и перезапустить API:
```bash
# в окне сервера Ctrl+C и:
export API_HTTP_PORT=8080 JWT_SECRET=devsecret ADMINS=$USER_ID
go build ./cmd/api && ./api
```

Проверка /stats/overview с правами:
```bash
# вариант 1 — JWT
curl -s "http://localhost:8080/stats/overview?top=3" -H "Authorization: Bearer $TOKEN" | jq .
# вариант 2 — упрощённый: Bearer <user_id>
curl -s "http://localhost:8080/stats/overview?top=3" -H "Authorization: Bearer $USER_ID" | jq .
```

Проверка FTS-триггера:
```bash
docker exec -i microtube-db psql -U app -d microtube -c "
INSERT INTO app.videos(id,title,description,lang,tags,duration_s,uploaded_at)
VALUES (gen_random_uuid(),'FTS Trigger Check','Some text','en','{test}',10,now());
SELECT fts_tsv IS NOT NULL as has_fts FROM app.videos WHERE title='FTS Trigger Check';
"
```

### Авторизация
- Допустимые форматы заголовка:
  - Authorization: Bearer <JWT> (HS256, секрет `JWT_SECRET`)
  - Authorization: Bearer <user_id> (строка без точек)
- Для доступа к `/stats/overview` добавьте `user_id` в `ADMINS` и перезапустите API.
