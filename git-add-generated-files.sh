#! /bin/sh

find . -name \*.svg  |
    grep -v examples/_ |
    xargs git-add -f

git-add -f '*README.md'
