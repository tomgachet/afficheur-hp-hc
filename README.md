# afficheur-hp-hc

Petit outil CLI en Go + DuckDB qui indique la plage tarifaire actuelle (`HP` ou `HC`), sa periode, son heure de fin, le temps restant et la prochaine plage.

## Utilisation

```bash
go run .
```

Avec une date de test :

```bash
go run . -db :memory: -at "2026-06-17 14:30:00"
```

Avec une base DuckDB precise :

```bash
go run . -db ./conso_elec.duckdb
```

Forcer le rechargement du referentiel CSV :

```bash
go run . -db ./conso_elec.duckdb -reload-ref
```

Desactiver les couleurs ANSI :

```bash
go run . -no-color
```

## Tests

```bash
go test ./...
```

## Referentiel

Le referentiel source est `ressources/ref_time_slot.csv`. Il est importe dans DuckDB dans la table `reference.ref_time_slot` uniquement si cette table vient d'etre creee.

Si la table existe deja, elle est conservee telle quelle. Pour remplacer son contenu par celui du CSV, utiliser `-reload-ref`, ce qui fait un `TRUNCATE TABLE reference.ref_time_slot` puis un nouvel import.

Il utilise la convention DuckDB `strftime('%w')` :

- dimanche = `0`
- lundi = `1`
- mardi = `2`
- mercredi = `3`
- jeudi = `4`
- vendredi = `5`
- samedi = `6`

Les heures de fin sont traitees en borne exclusive : une plage `00:00 -> 06:59` est calculee comme `[00:00 ; 07:00[`.
