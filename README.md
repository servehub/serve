#Манифест

Манифест файл в формате yaml, в котором в декларативном стиле описывается вся метаинформация, необходимая для автоматизации жизненного цикла компонента в процессах разработки: сборка, тестирование, деплой в разных средах, мониторинг и анализ работы и так далее. 

##Структура.

###Формат узлов.

Каждый узел документа может содержать директиву _match_ или ее короткий эквивалент „?“.
Указанная после имени узла директива _match_ говорит интерпретатору о специфичной интерпретации дочерних ключей узла. Имя каждого дочернего ключа становится паттерном матчинга для переменной, указанной после директивы _match_ в имени узла.
Только один дочерний узел может быть результатом паттерн-матчинга. В качестве приоритета паттернов используются следующие правила:
 - полное совпадение значения;
 - совпадение с регулярным выражением;
 - выражение "*";

Все дочернее дерево сматченного узла становится принимается как дочернее дерево узла с директивой _match_. Данные о директиве match после обработки удаляются.

Пример:
```
some ? {{ vars.env }}:
  q*: 
    name: q
  qa: qa
  live: live
```

Результат:
```
// result on vars.env == 'qa'
some: 
  name: q 
// because 'qa' matches 'q*' 
```

###Формат значений

В манифесте в значениях узлов можно использовать переменные. 
Синтаксис `https://golang.org/pkg/text/template/: "Привет {{ vars.variableName }}, как дела?".`

Все переменные окружения и аргументы, указанные в контексте интерпретации манифеста, помещаются в раздел _vars_. Там же для них можно указать значения по умолчанию. Через символ `|`  можно указывать функции-модификаторы с параметрами или без, первым параметром всегда будет предыдущее значение по конвееру обработки.

Пример:
```
vars:
  answer: 42
//а потом использовать:
meta:
  answer: "{{ vars.answer }}"
  home: "{{ vars.HOME|replace("\W","_")  }}"
  env: "{{ vars.env }}"
```

###Импорт манифестов

Для импорта деревьев из файла необходимо указать узел _file_ в дереве _import_ и указать путь к файлу YAML, узлы которого импортируются в текущий документ.
```
import:
  file: /etc/default/manifest.yml
```

При импорте деревьев может произойти перезапись одноименных дочерних узлов с узлами в импортируемых дереве. Приоритет перезаписи будут иметь описанные в текущем документе.

Пример:

```
// file_1.yml
var: other
var-1: var 1
```
```
// file_2.yml
var: var 1
import:
 - file: file_1.yml
```

Результат:
```
var: var 1
var-1: var 1
```

##Правила объединения
Правило слияния map.
При слиянии перезаписываются соответствующие узлы. Результат зависит от того, какой из этих элементов является базовым при слиянии. Базовым определяется элемент, который следует первым в порядке потока документа.

Пример:
```
var:
  var-1: "var 1"
  var-2: "var 2"

var:
  var-1: "new var 1"
  var-3: "var-3"

```

Результат:
```
var:
  var-1: "new var 1"
  var-2: "var 2"  
  var-3: "var-3"
```

Правило слияния списков.

**Правило 1**
При слиянии двух списков результирующий список определяется как соединение двух списков.

**Правило 2**
При слиянии списка и map результат зависит от того, какой из этих элементов является базовым при слиянии. Базовым определяется элемент, который следует первым в порядке потока документа.

**Правило 2.1**
Если базовым элементом является map, то каждый ключ этой map используется как базовый для одноименных ключей каждого элемента в списке. В итоге получается список определенный от оригинального, каждый элемент которого расширен базовой map.

Пример 2.1.1
```
import:
  stages:
    debian:
      repo: some
    zip:
      compression: deflate

stages:
  - sh: false
  - debian:
      name: one
  - zip:
      name: two
  - debian:
      name: three
```

Результат:
```
stages:
  - sh: false
  - debian:
      repo: some
      name: one
  - zip:
      compression: deflate
      name: two
  - debian:
      repo: some
      name: three
```

Пример:
```
import:
  stages:
    type: generic

stages:
  - debian: yes
    type: debian
  - zip: yes
  - type: other
```

Результат:
```
stages:
  - debian: yes
    type: debian
  - zip: yes
    type: generic
  - type: other
```

**Правило 2.2**
Если базовым элементом является список, то результат определяется как оригинальный map и список игнорируется.

Пример
```
import:
  stages:
    - sh: echo 1
    - sh: echo 2

stages:
  sh: echo 3
```

Результат
```
stages:
  sh: echo 3
```


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
При указании этого флага serve только распарсит манифест и выведет название плагина и параметры запуска. Если при этом не указать плагин, то будет выведен весь манифест. Полезно для отладки

 * `--plugin-data='{"some":"json"}'`
Можно вызвать конкретный плагин, передав ему на вход ранее сформированный json со всеми необходимыми параметрами. Тогда serve не будет искать и парсить манифест, а просто вызовет указанный плагин с переданными параметрами. 

Утилиту serve можно запустить указав плагин, которому будет передана часть дерева, определенная для плагина и управление. serve осуществляет поиск плагинов по следующему правилу:
 - поиск начинается с узла, указанного в качестве `<plugin-name>`;
 - узел является элементом списка, в этом случае осуществляется поиск плагина с именем составленным из имени узла и родительских узлов через разделитель ".".
Если плагин в вернет неуспешный статус выполнения, то следующие плагины в запускать не будут.
При запуске плагина содержимое одноименного узла передается в виде простого дерева на вход плагину.

Утилита serve выполняет поиск плагинов:
 - в папке /etc/serve/plugins/;
 - в реестре плагинов, с которыми она скомпилирована.


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

### db.create.postgresql 

Создает клон или удаляет БД Postgresql. Параметры для клонирования
```
purge: true/false
ssh-user: user-name
target: target-dbname
source: source-dbname
host: dbserver-host
```
Если в параметр `purge=true` будет произведена попытка удаления БД.
