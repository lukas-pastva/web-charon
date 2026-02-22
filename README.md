# Charon - Webova aplikace

Charon je webova aplikace pro spravu clanku, galerii a komentaru s administracnim rozhranim.

## Pristup na stranky

Verejna cast webu je dostupna na hlavni adrese (`/`). Zde najdete:

- **Clanky** — seznam publikovanych clanku na `/articles`, detail clanku na `/articles/{slug}`
- **Galerie** — prehled galerii na `/gallery`, detail galerie na `/gallery/{slug}`
- **Komentare** — u clanku mohou navstevnici pridavat komentare

## Prihlaseni do administrace

Administrace je dostupna na adrese `/admin/login`.

Pri prvnim spusteni aplikace se automaticky vytvori ucet s prezdivkou **admin**. Heslo se nastavi pomoci promenne prostredi `ADMIN_PASSWORD`. Pokud neni nastavena, pouzije se vychozi heslo `admin`.

1. Otevrete `/admin/login`
2. Zadejte prezdivku (nickname) a heslo
3. Po prihlaseni budete presmerovani na dashboard

## Administrace

Po prihlaseni mate pristup k nasledujicim sekcim:

### Dashboard

Prehled zakladnich statistik — pocet clanku, galerii a cekajicich komentaru.

### Clanky

- **Seznam clanku** — `/admin/articles`
- **Novy clanek** — kliknete na "New Article", vyplnte nazev, slug, obsah, vyber obalky a stav publikace
- **Uprava clanku** — kliknete na "Edit" u prislusneho clanku
- **Smazani clanku** — kliknete na "Delete" (s potvrzenim)

### Galerie

- **Seznam galerii** — `/admin/galleries`
- **Nova galerie** — vyplnte nazev, slug, popis a volitelne prirazeni ke clanku
- **Nahravani obrazku** — v editaci galerie nahrajte obrazky pres formular
- **Smazani obrazku** — kliknete na "Delete" u obrazku

### Komentare

- **Seznam komentaru** — `/admin/comments`
- **Schvaleni** — kliknete na "Approve" pro zverejneni komentare
- **Smazani** — kliknete na "Delete" pro odstraneni komentare

### Nastaveni

- **Povoleni komentaru** — zapnete nebo vypnete moznost pridavani komentaru na webu

### Sprava uzivatelu (pouze pro administratory)

Sekce `/admin/users` je pristupna pouze uzivatelum s opravnenim administratora.

- **Seznam uzivatelu** — prehled vsech uzivatelu systemu
- **Novy uzivatel** — vyplnte prezdivku (login), jmeno, prijmeni, heslo a zvolte, zda ma byt administrator
- **Uprava uzivatele** — zmente udaje, heslo je volitelne (pokud ho nevyplnite, zustane puvodni)
- **Smazani uzivatele** — odstraneni uzivatele ze systemu

### Profil

Kazdy prihlaseny uzivatel si muze upravit svuj profil na `/admin/profile`:

- **Jmeno a prijmeni** — upravte sve osobni udaje
- **Heslo** — zadejte nove heslo (pokud ho chcete zmenit)
- Prezdivku (login) nelze menit

## Promenne prostredi

| Promenna | Popis | Vychozi hodnota |
|---|---|---|
| `DB_HOST` | Adresa databazoveho serveru | `localhost` |
| `DB_PORT` | Port databaze | `3306` |
| `DB_USER` | Uzivatel databaze | `charon` |
| `DB_PASSWORD` | Heslo k databazi | _(prazdne)_ |
| `DB_NAME` | Nazev databaze | `charon` |
| `STORAGE_PATH` | Cesta pro ukladani souboru | `/data/uploads` |
| `PUBLIC_DOMAIN` | Verejna domena | `localhost` |
| `ADMIN_PASSWORD` | Heslo pro pocatecniho administratora | `admin` |
| `PORT` | Port, na kterem aplikace nasloucha | `8080` |
