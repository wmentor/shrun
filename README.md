# ShRun

##  Установка

Для сборки и использования shrun нужно иметь установленный docker, git, make и go (1.20.5). 

В переменную PATH нужно добавить путь к директории $GOPATH/bin.

Клонируем репозиторий и собираем shrun.

```bash
git clone git@github.com:wmentor/shrun.git
cd shrun
make
```

Для обновления будет необходимо подтянуть изменения из репозитория и сделать make. Важно учитывать тот момент, что
после обновления возможно нужно будет также выполнить команды *init*/*build*.

## Инициализация

```
shrun init
```

В конфигурационной директории создает Dockerfile-ы, sdmspec.json, rc.local. В качестве директории по умолчанию используется *~/.shrun* (если ее нет, 
то она будет создана при первом запуске). Для того чтобы поменять директорию по умолчанию нужно задать переменную окружения *SHRDM_CONFIG_DIR*.

Для получения информации о всех параметрах команды вызовите ее с ключем *-h*.

## Получения необходимых образов докера

```
shrun pull
```

В результате будут синхронизованы имеджи для go, ubuntu, postgres.

## Сборка образа Shardman

```
shrun build --build-basic --build-pg --build-gotpc
```

Для докеров используется каталог ~/build (может быть ссылкой). Если нужно его сменить, то стоит задать переменную окружения *SHRDM_DATA_DIR*.
В *build* должны быть три каталога: *shardman*, *shardman-utils*, *go-tpc*.

Если не задать ключи --build-basic и --build-pg, то будет только пересобран образ с новой обвязкой на базе последней сборки постгреса.

*--build-basic* нужен для пересборки всех базовых образов, которые используются для сборки постгреса, но сам постгрес при этом не собирается.

*--build-pg* указывает на то, что нужно пересобрать постгрес.

*--build-gotpc* указывает на то, что нужно собрать образ для использования go-tpc.

Обвязка пересобирается при любой конфигурации ключей.

## Запуск кластера

```
shrun start --nodes count [--update|-u] [--force|-f] [--mount-data] [--shell] [--grafana]
```

Запускает кластер из заданного числа нод (по умолчанию выполняется shardmanctl init + shardmanctl nodes add). 

Если ноды не нужно добавлять в кластер,  то нужно добавить флаг --skip-node-add.

Собрать следующий кластер можно будет теперь только после остановки (даже если сборка прошла неуспешно). Если добавить опцию *--force|-f*,
то в этой ситуации старый кластер будет остановлен и запущен новый. Если нужно перед запуском пересобрать утилиты, то нужно добавить
опцию *--update|-u* .

Флаг *--shell* говорит о том, что после добавления нод сразу нужно подключится к первой ноде.

Флаг *--grafana* подключает grafana/prometheus для кластера.

В случае успешного запуска в build-каталоге будет создана директория /mntdata, которая будет подмонтирована ко всем запущенным контейнерам.

Если добавить флаг *--mount-data*, в каталоге *<build>/pgdata* будет создан каталог *<container_name>*, который будет подмонтирована к каталогу
данных Shardman. После остановки через команду *stop*, этот каталог будет удален.

## Запуск дополнительных нод

```
shrun nodes add -n count [--mount-data]
```

Поднимает еще заданное число нод Shardman. При этом в кластер они автоматом не добавляются. Команда может
быть использована только после *shrun start*.

Если добавить флаг *--mount-data*, в каталоге *<build>/pgdata* будет создан каталог *<container_name>*, который будет подмонтирована к каталогу
данных Shardman. После остановки через команду *stop*, этот каталог будет удален.

## Удаление заданного числа нод Shardman

```
shrun nodes rm -n count
```

Удаляет заданное число нод Shardman или всех, если нод меньше чем заданное число. Если все ноды Shardman удалены
команда *stop* не выполняется т.к. еще остаются живые etcd-ноды.

## Подключение к конкретное ноде

```
shrun shell -n node -u user
```

Коннектится к заданной ноде из-под указанного пользователя (по дефолту используется пользователь *postgres* и нода *shrn1*).

## Подключение к базе данных на конкретной ноде

```
shrun psql -n node [-p port]
```

Коннектится к базе данных на указанной ноде (по дефолту используется нода *shrn1*). Порт (если не задан через -p), логин и пароль
берутся из *sdmspec.json* в директории конфигов.

## Остановка всех нод/сетей

```
shrun stop
```

## Удаление всех образов и очистка кэша сборки

```
shrun clean [--force|-f]
```
Перед очисткой происходит остановка всех нод, а также удаление использованной сети. Если задан флаг *--force|-f*, то образы удаляются принудительно.

## Генерация документации по Shardman

```
shrun doc
```

Команда сгенерирует документацию по шардману и напечатает каталог, в который она была сохранена.

## Запуск контейнер билдера

```
shrun gobuilder [--rebuild|-r]
```

Команда запускает контейнер *gobuilder*, в котором установлены все необходимые утилиты для сборки *shardman-utils* и подмонтированы
все нужные директории. Если задан флаг *--rebuild|-r*, то перед запуском будет пересобран образ контейнера.

## Запуск нагрузочный тестов через go-tpc

```
shrun gotpc
```

Команда запускает контейнер с установленной утилитой go-tpc для проведения нагрузочного тестирования кластера.
После поднятия контейнера сразу коннектится к нему и можно использовать утилиту go-tpc.
Если задан флаг *--rebuild|-r*, то перед запуском будет пересобран образ контейнера.
Важный момент, перед этой командой должен быть запущен кластер командой *start*.

Находясь в контейнере, задаем команду чтобы подготовить данные (если у нас есть три ноды shrn1,shrn2,shrn3):

```
go-tpc tpcc prepare -d postgres -U postgres -p 12345  -D postgres -H shrn1,shrn2,shrn3 -P 5432,5432,5432 \
       --conn-params sslmode=disable --partition-type 5 --warehouses 16 --parts 16 -T 16 --no-check
```

После подготовки запускаем тест:

```
go-tpc tpcc run -d postgres -U postgres -p 12345  -D postgres -H shrn1,shrn2,shrn3 -P 5432,5432,5432 \
       --conn-params sslmode=disable --partition-type 5 --warehouses 16 --parts 16 -T 32 --time 10m --ignore-error
```
