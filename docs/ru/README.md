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
  <a href="../../README.md">English</a> |
  <span>Русский</span>
</p>

* Статус проекта: Реализован концепт

[pgfc] - golang библиотека для вызова хранимых функций postgresql без создания структур golang, описывающих их аргументы и результат.

Эта библиотека представляет собой вариант доказательства гипотезы "Большую часть операций с БД можно свести к запросу вида `SELECT * FROM function(args)`" применительно к postgresql и предназначена для создания универсальной прослойки между прикладными (SQL) разработчиками и разработчиками фронтендов.

## Алгоритм

```go
db := pgfc.NewServer(dsn)
args := map[string]interface{}{
	"arg1": "name",
	"arg2": 1
}
rv, err := db.Call("method", args)
```
Этот пример показывает, что код на go не должен заранее знать, какие есть функции, какие у них аргументы и результаты. Это позволяет предоставить универсальный доступ к хранимым функциям postgresql для серверных шаблонов и внешних клиентов. Пример такого решения - [gin-pgfc].

## Требования к БД

Список хранимых функций, описание их аргументов и результатов можно получить из БД простыми SQL, но при этом возникают вопросы:

* как отличить служебную функцию от доступной извне?
* как, не меняя клиентов, изменить имя вызываемой извне функции?
* куда положить, для документации, комментарии к аргументам функций?
* куда положить, для документации, пример вызова функции?

Вариант ответа на эти вопросы - требование наличия в БД функций

	InDefFunc     string `long:"db_indef" default:"func_args" description:"Argument definition function"`
	OutDefFunc    string `long:"db_outdef" default:"func_result" description:"Result row definition function"`
	IndexFunc     string `long:"db_index" default:"index" description:"Available functions list"`


	// SQLMethod is the SQL query for fetching method list
	// Results: err = rows.Scan(&r.Name, &r.Class, &r.Func, &r.Anno, &r.Sample, &r.Result, &r.IsRO, &r.IsSet, &r.IsStruct)
	SQLMethod = "select code, nspname, proname, anno, sample, result, is_ro, is_set, is_struct from %s($1)"
	// SQLInArgs is the SQL query for fetching method arguments definition
	// Results: err = rows.Scan(&r.Name, &r.Type, &r.Required, &r.Default, &r.Anno)
	SQLInArgs = "select arg, type, required, def_val, anno from %s($1)"
	// SQLOutArgs is the SQL query for fetching method results definition
	// Results: err = rows.Scan(&r.Name, &r.Type, &r.Anno)
	SQLOutArgs = "select arg, type, anno from %s($1)"

Пример реализации такого функционала  - pomasql/rpc

## Использование

TODO: example/simple.go
```

```

### См. также

* gin-pgfc - клей для gin-gonic
* apisite - фреймворк, использующий pgfc в шаблонах и внешних вызовах
* enfist - пример готового приложения

## Лицензия

Лицензия MIT (MIT), см. [LICENSE](LICENSE) (неофициальный перевод,
 [источник перевода](https://ru.wikipedia.org/wiki/%D0%9B%D0%B8%D1%86%D0%B5%D0%BD%D0%B7%D0%B8%D1%8F_MIT), [оригинал лицензии](../../LICENSE)).

Copyright (c) 2018 Алексей Коврижкин <lekovr+apisite@gmail.com>
