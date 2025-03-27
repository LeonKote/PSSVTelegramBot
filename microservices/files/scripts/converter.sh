#!/bin/bash

INPUT="$1"
TMP="${INPUT}.tmp"

cp "$INPUT" "$TMP"

for path in $(yq e '.paths | keys | .[]' "$INPUT"); do
  echo "📂 Обрабатываем path: $path"

  # Получаем все методы в .paths.$path (post, put, etc.)
  for method in $(yq e ".paths.\"$path\" | keys | .[]" "$INPUT"); do
    if [[ "$method" != "post" && "$method" != "put" ]]; then
      continue
    fi

    echo "🔍 Метод: $method"

    # Проверим, есть ли parameters[].in == body
    has_body=$(yq e ".paths.\"$path\".$method.parameters[]? | select(.in == \"body\")" "$INPUT")
    if [[ -z "$has_body" ]]; then
      echo "⏭️ Нет параметра in: body, пропускаем"
      continue
    fi

    # Получим $ref
    ref=$(yq e ".paths.\"$path\".$method.parameters[] | select(.in == \"body\") | .schema.\$ref" "$INPUT")
    echo "📌 Найден $ref"

    # Соберем список ключей в исходном порядке
    keys=$(yq e ".paths.\"$path\".$method | keys | .[]" "$INPUT")

    # Удаляем весь блок метода
    yq e "del(.paths.\"$path\".$method)" -i "$TMP"

    # Создаем новый блок с сохранением порядка и без parameters
    TMPFILE=$(mktemp)
    {
      echo "paths:"
      echo "  $path:"
      echo "    $method:"
      for key in $keys; do
        if [[ "$key" == "parameters" ]]; then
          echo "      requestBody:"
          echo "        \$ref: '$ref'"
        else
          yq e ".paths.\"$path\".$method | {\"$key\": .\"$key\"}" "$INPUT" | sed 's/^/      /'
        fi
      done
    } > "$TMPFILE"

    # Объединяем новый блок с остальным YAML
    yq ea 'select(fileIndex == 0) * select(fileIndex == 1)' "$TMP" "$TMPFILE" > "${TMP}.tmp"
    mv "${TMP}.tmp" "$TMP"
    rm "$TMPFILE"

    echo "✅ Обновлён $method $path"
    echo
  done
done

cp "$TMP" "$INPUT"
#cp "$TMP" "autogen/docs/swagger1.yaml"
rm "$TMP"

echo "🏁 Всё готово! Изменения применены к: $INPUT"
