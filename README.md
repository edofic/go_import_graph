# go_import_graph

A simple command line tool to analyse Go package imports and render them as a PNG.

basic usage

```
go run github.com/edofic/go_import_graph -module="mymodule"
```

And an image will pop up.

Where "mymodule" is usually the name of the Go module you want to analyse. You
can also specify an arbitrary string here as this is only used for substring
matching when analysing imports.

Optionally you can specify folder other than current via `-module`.

By default a new file will be created in your tmp dir, you can override this by
specifying `-target`
