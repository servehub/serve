#!/usr/bin/env bash

# для serve на Go машинках нужен будет /etc/serve/config.yml,
# в котором будут адреса марафона, реестра пакетов и прочие конфиги

# - собираем пакет и загружаем артефакты в репозиторий (apt, task-registry, maven, etc)
serve app build --build-number '34' --branch 'master'

# - запускаем новую версию,
# - дожидаемся появления в консуле
serve app deploy --env 'qa' --build-number '34' --branch 'master'

# - находим сервис в консуле
# - добавляем ему роутинг-параметры чтобы он в nginx попал
# - удаляем предыдущую версию с таким же staging из консула
# - стопаем предыдущую версию с таким же staging в марафоне (через 3 минуты)
serve app route --env 'qa' --build-number '34' --branch 'master'   # --branch опционально поле, по-умолчанию master

serve app route --env 'live' --staging 'stage' --build-number '34'
serve app route --env 'live' --staging 'live' --build-number '34'


# скрипт-wrapper для регистрации
serve consul supervisor \
  --name 'forgame-api3-v1.0.34' \
  --build '1.0.34'
  --port 12073
  start bin/start -Xmx521m -Dinn.api.port=12073 ...
