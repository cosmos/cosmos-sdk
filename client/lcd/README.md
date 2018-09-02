# How to update API docs

Due to the rest handlers and related data structures are distributed in many sub-folds, currently there is no tool which can automatically extract all APIs information and generate API docs. So here we have to write APIs' docs manually.

## Steps
 
* Install the command line tool first.
```
go get github.com/rakyll/statik
```
* Directly Edit API docs: client/lcd/swaggerui/swagger.json

* Edit API docs within this [website](https://app.swaggerhub.com). Please refer to this [link](https://app.swaggerhub.com/help/index) for how to use the about website to edit API docs.

* Download swagger.json and replace the old swagger.json under client/lcd/swaggerui folds

* Regenerate statik.go file
```
statik -src=client/lcd/swaggerui -dest=client/lcd
```