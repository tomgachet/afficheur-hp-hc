# afficheur-hp-hc
Petit tableau de bord local en Go + DuckDB pour afficher la plage tarifaire électrique actuelle, la prochaine plage et les durées associées 

## Principe

- une base de données DuckDB référence les plages HP/HC
- une requête SQL retourne l’état actuel
- un petit service Go lit la requête et les met en forme en HTML
- la page HTML est visualisable sur le réseau local
