# afficheur-hp-hc

CLI en Go + DuckDB qui indique la plage tarifaire actuelle (`HP` ou `HC`), sa periode, son heure de fin, le temps restant et la prochaine plage.

## Utilisation 

Avec une base DuckDB en mémoire :

```bash
go run .
```

Avec une date de test :

```bash
go run . -db :memory: -at "2026-06-17 14:30:00"
```

Avec une base DuckDB précise:

```bash
go run . -db ./conso_elec.duckdb
```

Forcer le rechargement du réferentiel CSV en base :

```bash
go run . -db ./conso_elec.duckdb -reload-ref
```

Désactiver les couleurs ANSI :

```bash
go run . -no-color
```

## Tests

```bash
go test ./...
```

## Réferentiel

Le réferentiel source est `ressources/ref_time_slot.csv`. 
Les plages des tarification par défaut sont basées sur [l'offre flexibilité 2 saisons d'Enercoop] (https://www.enercoop.fr/notre-offre/flexibilite2saisons)

Il est importé dans DuckDB dans la table `reference.ref_time_slot` uniquement si cette table vient d'etre créee.

Si la table existe déjà, elle est conservée telle quelle. Pour remplacer son contenu par celui du CSV, utiliser `-reload-ref`, ce qui fait un `TRUNCATE TABLE reference.ref_time_slot` puis un nouvel import.

Il utilise la convention DuckDB `strftime('%w')` :

- dimanche = `0`
- lundi = `1`
- mardi = `2`
- mercredi = `3`
- jeudi = `4`
- vendredi = `5`
- samedi = `6`

Les heures de fin sont traitées en borne exclusive : une plage `00:00 -> 06:59` est calculée comme `[00:00 ; 07:00[`.
