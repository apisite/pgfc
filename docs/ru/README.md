<p align="center">
  <a href="../../README.md">English</a> |
  <span>Русский</span>
</p>

---

# pgfc
> golang пакет для вызова функций postgresql

[![GoCard][gc1]][gc2]
 [![GitHub Release][gr1]][gr2]
 [![GitHub code size in bytes][sz]]()
 [![GitHub license][gl1]][gl2]

[gc1]: https://goreportcard.com/badge/apisite/pgfc
[gc2]: https://goreportcard.com/report/github.com/apisite/pgfc
[gr1]: https://img.shields.io/github/release/apisite/pgfc.svg
[gr2]: https://github.com/apisite/pgfc/releases
[sz]: https://img.shields.io/github/languages/code-size/apisite/pgfc.svg
[gl1]: https://img.shields.io/github/license/apisite/pgfc.svg
[gl2]: LICENSE

<p align="center">
<a target="_blank" rel="noopener noreferrer" href="../src/arch.png"><img src="../src/arch.png" title="Архитектура проекта" style="max-width:100%;"></a>
</p>

* Статус проекта: Реализован концепт

[pgfc](https://github.com/apisite/pgfc) - golang package для выполнения в Postgresql запросов вида `SELECT * FROM function(...)` в случае, когда список и сигнатуры функций заранее неизвестны.
Проект имеет целью создание универсальной прослойки между прикладными (SQL) разработчиками и разработчиками фронтендов.

## Использование

### Postgresql

В БД должны быть созданы функции (код из конфигурации pgfc):

	InDefFunc     string `long:"db_indef" default:"func_args" description:"Argument definition function"`
	OutDefFunc    string `long:"db_outdef" default:"func_result" description:"Result row definition function"`
	IndexFunc     string `long:"db_index" default:"index" description:"Available functions list"`

Эти функции используются для загрузки метаданных:

	// SQLMethod is the SQL query for fetching method list
	// Results: err = rows.Scan(&r.Name, &r.Class, &r.Func, &r.Anno, &r.Sample, &r.Result, &r.IsRO, &r.IsSet, &r.IsStruct)
	SQLMethod = "select code, nspname, proname, anno, sample, result, is_ro, is_set, is_struct from %s($1)"
	// SQLInArgs is the SQL query for fetching method arguments definition
	// Results: err = rows.Scan(&r.Name, &r.Type, &r.Required, &r.Default, &r.Anno)
	SQLInArgs = "select arg, type, required, def_val, anno from %s($1)"
	// SQLOutArgs is the SQL query for fetching method results definition
	// Results: err = rows.Scan(&r.Name, &r.Type, &r.Anno)
	SQLOutArgs = "select arg, type, anno from %s($1)"

Пример реализации такого функционала -  [pomasql/rpc](https://github.com/pomasql/rpc)

### Golang

```go
db := pgfc.NewServer(dsn)
args := map[string]interface{}{
	"arg1": "name",
	"arg2": 1
}
rv, err := db.Call("method", args)
```

См. также: [gin-pgfc](https://github.com/apisite/gin-pgfc).

TODO: example/simple.go

## Требования к БД

Наличие в БД функций для метаданных является вариантом ответов на следующие вопросы:

* как отличить служебную функцию от доступной извне?
* как, не меняя клиентов, изменить имя вызываемой извне функции?
* куда положить, для документации, комментарии к аргументам функций?
* куда положить, для документации, пример вызова функции?


### См. также

* [gin-pgfc](https://github.com/apisite/gin-pgfc) - клей для gin-gonic
* [apisite](https://github.com/apisite/apisite) - фреймворк, использующий pgfc в шаблонах и внешних вызовах
* [enfist](https://github.com/apisite/app-enfist) - пример готового приложения

## Лицензия

Лицензия MIT (MIT), см. [LICENSE](LICENSE) (неофициальный перевод,
 [источник перевода](https://ru.wikipedia.org/wiki/%D0%9B%D0%B8%D1%86%D0%B5%D0%BD%D0%B7%D0%B8%D1%8F_MIT), [оригинал лицензии](../../LICENSE)).

Copyright (c) 2018 Алексей Коврижкин <lekovr+apisite@gmail.com>
