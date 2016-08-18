# serve

Утилита для парсинга манифеста и запуска плагинов

    serve <plugin-name> args...
    
Поддерживаемые флаги:

 * `--manifest=./manifest.yml`
Путь до файла манифеста, по-умолчанию ищется в текущей дериктории.
    
 * `--var name=value`
Параметры, которые нужно передать в манифест. В манифесте будут доступны так {{ var.name }}. Можно указать несколько. 
Например так: `serve deploy --var env=qa --var build-number=34 --var branch=feature-superman`

 * `--dry-run`
При указании этого флага serve только распарсит манифест и выведет название плагина и параметры запуска. Полезно для отладки

 * `--plugin-data='{"some":"json"}'`
Можно вызвать конкретный плагин, передав ему на вход ранее сформированный json со всеми необходимыми парамерами. Тогда serve не будет искать и парсить манифест, а просто вызовет указанный плагин. 

## Плагины

### build.sh
Просто вызывает любую sh команду.
```
{
  "sh": "echo 'hello world!'"
}
```

### build.sbt-pack
Вызывает sbt с параметрами:
    
    sbt ';set version := "%s"' clean test pack
    
```
{
  "version": "1.3.42"
}
```

### build.marathon
Упаковывает `source` директорию в tar.gz архив и заливает через webdav в marathon task-registry по указанному в `registry-url` урлу.
```
{
  "registry-url": "http://mesos1-q.qa.inn.ru/task-registry/kidzania/kidzite-api/kidzite-api-v2.1.34.tar.gz",
  "source": "target/pack"
}
```

### build.debian
Создает deb-пакет и заливает его в apt репозиторий. Используются inn-ci-tools скрипты.
```
{
  "build-number": "34",
  "category": "kidzania",
  "ci-tools-path": "/var/go/inn-ci-tools",
  "cron": "",
  "daemon": "$APPLICATION_PATH/bin/start",
  "daemon-args": "--port=$PORT1 --env=${ENV_NAME}",
  "daemon-port": 9040,
  "daemon-user": "innova",
  "depends": "python3-setuptools, python3-dev, python3-pip",
  "description": "Kidzania CMS v1.6.0",
  "distribution": "unstable",
  "init": "debian-way",
  "install-root": "/local/innova/www-versions",
  "maintainer-email": "bamboo@inn.ru",
  "maintainer-name": "Continuous Integration",
  "make-pidfile": "yes",
  "name": "inn-kidzania-cms",
  "package": "inn-kidzania-cms-v1.6.34",
  "service-owner": "innova",
  "stage-counter": "0",
  "version": "1.6"
}
```

### deploy.marathon
Запускает приложение в марафоне. Оборачивает `cmd` в специальный скрипт-враппер, который регистрирует запущенный сервис в консуле в реестре сервисов. 
```
{
  "app-name": "kidzania/kidzite-api-v2.1.34",
  "cmd": "bin/start",
  "constraints": "kidz:true",
  "consul-host": "mesos1-q.qa.inn.ru",
  "cpu": 0.1,
  "environment": {
    "ENV": "qa",
    "MEMORY": "512",
    "SERVICE_NAME": "kidzite-api",
    "SERVICE_VERSION": "2.1.34"
  },
  "instances": 1,
  "marathon-host": "mesos1-q.qa.inn.ru",
  "mem": 512,
  "package-uri": "http://mesos1-q.qa.inn.ru/task-registry/kidzania/kidzite-api/kidzite-api-v2.1.34.tar.gz"
}
```

### release
Обновляет маршруты для http-роутинга в консуле. На базе этих данных consul-template автоматически сгенерит конфиг nginx
В параметре `--var route='{"stage":"staging"}'` можно передавать доп. параметры для роутинга. 

```
{
  "consul-host": "mesos1-q.qa.inn.ru",
  "full-name": "kidzania/kidzite-api-v2.1.34",
  "name-prefix": "kidzania/kidzite-api-v",
  "outdated": {
    "marathon": {
      "marathon-host": "mesos1-q.qa.inn.ru"
    }
  },
  "route": "",
  "routes": [
    {
      "host": "api2.kidzite.qa",
      "location": "/"
    }
  ]
}
```
