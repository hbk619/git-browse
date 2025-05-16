#!/bin/bash
set -e

latest_tag=$(git describe --tags --abbrev=0)

echo "## What's changed"
pr_regex="/#[0-9]+/"
while IFS=£ read -r commit_body author subject; do    
    if [[ $subject =~ Merge\ pull\ request\ #([0-9]+) ]]; then
        pr_number=${BASH_REMATCH[1]}
        echo "* $commit_body by $author in https://github.com/hbk619/gh-peruse/pull/$pr_number"
    fi
done <<< "$(git log --merges --pretty=format:"%b£%an£%s" "$latest_tag"..HEAD)" 

echo ""
echo ""
echo "**Full Changelog**: https://github.com/hbk619/gh-peruse/compare/$latest_tag..$1"