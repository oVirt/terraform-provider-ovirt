#!/bin/bash

export NOLINT=1
go generate >/tmp/gogenerate.output 2>/tmp/gogenerate.output
if [ $? -ne 0 ]; then
  echo -e "::group::\e[0;31m❌ Go generate failed.\e[0m"
  cat /tmp/gogenerate.output
  echo "::endgroup::"
  exit 1
fi

echo -e "::group::\e[0;32m✅ Go generate succeeded.\e[0m"
cat /tmp/gogenerate.output
echo "::endgroup::"

git diff >/tmp/gogenerate.diff
if [ "$(cat /tmp/gogenerate.diff | wc -l)" -ne 0 ]; then
  echo -e "::group::\e[0;31m❌ Git changes after go generate.\e[0m"
  echo "The following is the diff of files:"
  cat /tmp/gogenerate.diff
  echo "::endgroup::"
  echo -e "\e[0;31m⚙ Please run go generate before pushing your changes.\e[0m"
  exit 1
fi

echo -e "::group::\e[0;32m✅ No changes after go generate.\e[0m"
echo "::endgroup::"
