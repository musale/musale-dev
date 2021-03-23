hugo -t hoyt
cd public
git add .
git commit -m "Rebuilding the site on %date% -%time%"
git push origin master