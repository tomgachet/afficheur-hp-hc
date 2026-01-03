# Analyse de la consommation électique du foyer dans une base de données DuckDB

## Contexte

Avec un compteur électique nouvelle génération, on peut récupérer auprès de son fournisseur d'énergie le détail de sa consommation électrique par plage de demi-heure.
En général, ce même founisseur propose sur son site internet plusieurs visualisations mettant en forme ces données.

Etant abonné au tarif de base sans heures pleines (HP) et heures creuses (HC), je souhaitais pouvoir étudier les différentes offres en HP/HC de mon fournisseur d'électricité en confrontant ma consommation aux différentes plages et tarifs appliqués.

Cette étude me permet également de ma familiariser avec la base de données DuckDB, qui est parfaite pour ce type d'analyse (légère, rapide et serverless)

## Les données

Fournisseur : Enercoop

D'un simple clic dans mon espace client, je peux télécharger un *.csv* avec une colonne d'horodatage (toutes les demi-heures) et la consommation électrique réalisée pendant la demi-heure.

## Intégration à DuckDB

J'utilise la CLI de DuckDB pour la création de la base de données initiale et l'intégration des données de consommation à partir du fichier téléchargé. 

Je connecte ensuite cette base de données à l'IDE DBeaver pour intéragir plus directement en SQL.
