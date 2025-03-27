#!/bin/bash

# Путь к директории с документацией
DOCS=$1
PATHS=$2
SWAG=github.com/swaggo/swag/cmd/swag@latest

# Генерируем swagger спецификацию на основе анотаций
go run $SWAG init --parseDependency -d $PATHS -o $DOCS -g main.go
# Подготавливаем данные для подключения в docs.go
cat ./README.md | sed 's/"/\\"/g' > $DOCS/README.md
awk -F/ '{print "\"" $0 "\\n\"+"}' $DOCS/README.md > $DOCS/README.data
sed -i -e '1 s/^/ Description:/;' $DOCS/README.data
echo "\"\"," >> $DOCS/README.data
# Подключаем README.md в docs.go
sed -e '/%README_FILE%/{' -e 'r $DOCS/README.data' -e 'd' -e '}' $DOCS/docs.go > $DOCS/tmp.go
mv $DOCS/tmp.go $DOCS/docs.go
sed -i -e '/LeftDelim/d' $DOCS/docs.go
sed -i -e '/RightDelim/d' $DOCS/docs.go
sed -i -e 's/version:\ \"2\.0\"/version:\ \"3.0\"/g' $DOCS/swagger.yaml
sed -i -e '/^ *parameters:/,/^[^ ]/{ /^ *type:/ s/type:/schema:\n          type:/; }' $DOCS/swagger.yaml