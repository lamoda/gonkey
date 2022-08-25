# v1.18.3 (Thu Aug 25 2022)

#### 🐛 Bug Fix

- Fix finalizing allure reports. [#179](https://github.com/lamoda/gonkey/pull/179) ([@vitkarpenko](https://github.com/vitkarpenko))
- With-db-example .PHONY fix. [#172](https://github.com/lamoda/gonkey/pull/172) ([@vitkarpenko](https://github.com/vitkarpenko))
- feature: redis fixtures support [#176](https://github.com/lamoda/gonkey/pull/176) (aleksandr.nemtarev@lamoda.ru [@anemtarev](https://github.com/anemtarev))

#### 📝 Documentation

- Added JSON-Schema and how to set it up in readme [#171](https://github.com/lamoda/gonkey/pull/171) ([@leorush](https://github.com/leorush))

#### Authors: 4

- [@anemtarev](https://github.com/anemtarev)
- Alexander Nemtarev (aleksandr.nemtarev@lamoda.ru)
- Lev ([@leorush](https://github.com/leorush))
- Vitaly Karpenko ([@vitkarpenko](https://github.com/vitkarpenko))

---

# v1.18.3 (Mon Aug 08 2022)

#### Enhancement

- Redis fixtures support
- Custom loader support if using gonkey as a library with a FixtureLoader configuration attribute

#### Authors: 1

- Alexander Nemtarev [#178](https://github.com/lamoda/gonkey/pull/176) ([@anemtarev](https://github.com/anemtarev))

# v1.18.2 (Fri Jul 08 2022)

#### 🐛 Bug Fix

- Aerospike fixtures support [#168](https://github.com/lamoda/gonkey/pull/168) ([@vitkarpenko](https://github.com/vitkarpenko))

#### Authors: 1

- Vitaly Karpenko ([@vitkarpenko](https://github.com/vitkarpenko))

---

# v1.18.1 (Wed Jun 22 2022)

#### 🐛 Bug Fix

- Use regexp matching inside database response checks [#166](https://github.com/lamoda/gonkey/pull/166) ([@Cdayz](https://github.com/Cdayz))

#### Authors: 1

- Nikita Tomchik ([@Cdayz](https://github.com/Cdayz))

---

# v1.18.0 (Tue Jun 21 2022)

#### 🚀 Enhancement

- Add dbChecks format, for run multiply sql queries in one test [#164](https://github.com/lamoda/gonkey/pull/164) ([@Cdayz](https://github.com/Cdayz))

#### Authors: 1

- Nikita Tomchik ([@Cdayz](https://github.com/Cdayz))

---

# v1.17.0 (Wed Jun 08 2022)

#### 🚀 Enhancement

- Add `template` mock strategy for use incoming request inside mock responses [#162](https://github.com/lamoda/gonkey/pull/162) ([@Cdayz](https://github.com/Cdayz))

#### 📝 Documentation

- Added table of contents. Also fixed some markdownlint issues in README. [#160](https://github.com/lamoda/gonkey/pull/160) (vitaly.karpenko@lamoda.ru)

#### Authors: 2

- Nikita Tomchik ([@Cdayz](https://github.com/Cdayz))
- Vitaly Karpenko ([@vitkarpenko](https://github.com/vitkarpenko))

---

# v1.16.1 (Wed Jun 01 2022)

#### 🐛 Bug Fix

- feat: add ? to query [#159](https://github.com/lamoda/gonkey/pull/159) ([@sashamelentyev](https://github.com/sashamelentyev))

#### Authors: 1

- Sasha Melentyev ([@sashamelentyev](https://github.com/sashamelentyev))

---

# v1.16.0 (Thu May 26 2022)

#### 🚀 Enhancement

- New: ignore db ordering in dbResponse feature [#154](https://github.com/lamoda/gonkey/pull/154) (lev.marder@lamoda.ru)

#### Authors: 1

- Lev Marder ([@ikramanop](https://github.com/ikramanop))

---

# v1.15.0 (Thu May 12 2022)

#### 🚀 Enhancement

- New: regexp in query matching [#132](https://github.com/lamoda/gonkey/pull/132) (lev.marder@lamoda.ru)

#### 🏠 Internal

- Bump github.com/stretchr/testify from 1.7.0 to 1.7.1 [#156](https://github.com/lamoda/gonkey/pull/156) ([@dependabot[bot]](https://github.com/dependabot[bot]))
- #133 | fix data race [#155](https://github.com/lamoda/gonkey/pull/155) ([@architectv](https://github.com/architectv))

#### Authors: 3

- [@dependabot[bot]](https://github.com/dependabot[bot])
- Alexey Vasyukov ([@architectv](https://github.com/architectv))
- Lev Marder ([@ikramanop](https://github.com/ikramanop))

---

# v1.14.1 (Wed May 11 2022)

#### 🐛 Bug Fix

- new: add comparisonParams to BodyMatches for Json and XML [#117](https://github.com/lamoda/gonkey/pull/117) (Alexey.Tyuryumov@acronis.com [@Alexey19](https://github.com/Alexey19))

#### Authors: 2

- [@Alexey19](https://github.com/Alexey19)
- Alexey Tyuryumov (Alexey.Tyuryumov@acronis.com)

---

# v1.14.0 (Fri Feb 25 2022)

#### 🚀 Enhancement

- new: basedOnRequest strategy implemented [#130](https://github.com/lamoda/gonkey/pull/130) (eliseeviam@gmail.com [@eliseeviam](https://github.com/eliseeviam))

#### Authors: 2

- Aleksei Eliseev ([@eliseeviam](https://github.com/eliseeviam))
- eliseeviam (eliseeviam@gmail.com)

---

# v1.13.2 (Mon Jan 24 2022)

#### 🐛 Bug Fix

- chore(deps): Upgrade github.com/tidwall/gjson and set dependabot [#121](https://github.com/lamoda/gonkey/pull/121) ([@sashamelentyev](https://github.com/sashamelentyev))

#### Authors: 1

- Sasha Melentyev ([@sashamelentyev](https://github.com/sashamelentyev))

---

# v1.13.1 (Wed Dec 29 2021)

#### 🐛 Bug Fix

- fix: error in documentation [#116](https://github.com/lamoda/gonkey/pull/116) (Alexey.Tyuryumov@acronis.com [@Alexey19](https://github.com/Alexey19))

#### Authors: 2

- [@Alexey19](https://github.com/Alexey19)
- Alexey Tyuryumov (Alexey.Tyuryumov@acronis.com)

---

# v1.13.0 (Tue Dec 21 2021)

#### 🚀 Enhancement

- Add `skipped`, `broken` and `focus` statuses for tests definitions & export statuses in allure report's [#115](https://github.com/lamoda/gonkey/pull/115) ([@Cdayz](https://github.com/Cdayz))

#### Authors: 1

- Nikita Tomchik ([@Cdayz](https://github.com/Cdayz))

---

# v1.12.0 (Thu Dec 09 2021)

#### 🚀 Enhancement

- Allow passing variables to dbQuery and dbResponse [#114](https://github.com/lamoda/gonkey/pull/114) ([@Cdayz](https://github.com/Cdayz))

#### Authors: 1

- Nikita Tomchik ([@Cdayz](https://github.com/Cdayz))

---

# v1.11.1 (Tue Dec 07 2021)

#### 🐛 Bug Fix

- Write paths to test files in test output [#113](https://github.com/lamoda/gonkey/pull/113) ([@Cdayz](https://github.com/Cdayz))

#### Authors: 1

- Nikita Tomchik ([@Cdayz](https://github.com/Cdayz))

---

# v1.11.0 (Mon Dec 06 2021)

#### 🚀 Enhancement

- added bodyMatchesText requestConstraint with example test [#110](https://github.com/lamoda/gonkey/pull/110) (l.yarushin@timeweb.ru [@leorush](https://github.com/leorush))

#### Authors: 2

- Lev ([@leorush](https://github.com/leorush))
- Lev Yarushin (l.yarushin@timeweb.ru)

---

# v1.10.0 (Mon Dec 06 2021)

#### 🚀 Enhancement

- [BUG] fixed DbType param in console client [#112](https://github.com/lamoda/gonkey/pull/112) ([@chistopat](https://github.com/chistopat))

#### Authors: 1

- Yegor Chistyakov ([@chistopat](https://github.com/chistopat))

---

# v1.9.0 (Fri Dec 03 2021)

#### 🚀 Enhancement

- Use fixtures.Postgres as default DbType when run tests [#109](https://github.com/lamoda/gonkey/pull/109) ([@Cdayz](https://github.com/Cdayz))

#### Authors: 1

- Nikita Tomchik ([@Cdayz](https://github.com/Cdayz))

---

# v1.8.0 (Tue Nov 23 2021)

#### 🚀 Enhancement

- (service_mock): set server address from argument [#107](https://github.com/lamoda/gonkey/pull/107) ([@33r01b](https://github.com/33r01b))

#### Authors: 1

- Ruslan Samigullin ([@33r01b](https://github.com/33r01b))

---

# v1.7.1 (Sun Nov 21 2021)

#### 🐛 Bug Fix

- Fix race condition in service mock [#108](https://github.com/lamoda/gonkey/pull/108) ([@hibooboo2](https://github.com/hibooboo2))

#### Authors: 1

- James Jeffrey ([@hibooboo2](https://github.com/hibooboo2))

---

# v1.7.0 (Thu Oct 28 2021)

#### 🚀 Enhancement

- the ability to run afterRequestScript [#105](https://github.com/lamoda/gonkey/pull/105) (andrey.suchilov@lamoda.ru [@nofuture17](https://github.com/nofuture17))

#### Authors: 2

- Andrey Suchilov (andrey.suchilov@lamoda.ru)
- Suchilov Andrey ([@nofuture17](https://github.com/nofuture17))

---

# v1.6.0 (Fri Oct 15 2021)

#### 🚀 Enhancement

- the ability to use the schema in PostgreSQL is added [#102](https://github.com/lamoda/gonkey/pull/102) (andrey.suchilov@lamoda.ru [@nofuture17](https://github.com/nofuture17))

#### Authors: 2

- Andrey Suchilov (andrey.suchilov@lamoda.ru)
- Suchilov Andrey ([@nofuture17](https://github.com/nofuture17))

---

# v1.5.1 (Fri Sep 24 2021)

#### ⚠️ Pushed to `master`

- Minor fix in README.md ([@luza](https://github.com/luza))

#### 📝 Documentation

- doc: update Readme files - backslash in regexp should be escaped [#99](https://github.com/lamoda/gonkey/pull/99) ([@svzhl](https://github.com/svzhl))

#### Authors: 2

- [@svzhl](https://github.com/svzhl)
- Denis Girko ([@luza](https://github.com/luza))

---

# v1.5.0 (Wed Sep 22 2021)

#### 🚀 Enhancement

- new: fix algorithm for array compare [#98](https://github.com/lamoda/gonkey/pull/98) (Alexey.Tyuryumov@acronis.com [@Alexey19](https://github.com/Alexey19))

#### 🐛 Bug Fix

- new: support variables in mock definition [#94](https://github.com/lamoda/gonkey/pull/94) (Alexey.Tyuryumov@acronis.com [@Alexey19](https://github.com/Alexey19))

#### 📝 Documentation

- new: added missing $matchRegexp to examples [#95](https://github.com/lamoda/gonkey/pull/95) (Alexey.Tyuryumov@acronis.com [@Alexey19](https://github.com/Alexey19))

#### Authors: 2

- [@Alexey19](https://github.com/Alexey19)
- Alexey Tyuryumov (Alexey.Tyuryumov@acronis.com)

---

# v1.4.2 (Sat Sep 04 2021)

#### 🐛 Bug Fix

- mocks file reply strategy: show error when no file found by filepath [#96](https://github.com/lamoda/gonkey/pull/96) ([@NikolayOskin](https://github.com/NikolayOskin))

#### Authors: 1

- Nikolay Oskin ([@NikolayOskin](https://github.com/NikolayOskin))

---

# v1.4.1 (Thu Jun 24 2021)

#### 🐛 Bug Fix

- new: add support for "Host" http header [#91](https://github.com/lamoda/gonkey/pull/91) (Alexey.Tyuryumov@acronis.com [@Alexey19](https://github.com/Alexey19))

#### 📝 Documentation

- add docs in README: clear tables in fixtures [#92](https://github.com/lamoda/gonkey/pull/92) ([@NikolayOskin](https://github.com/NikolayOskin))
- fix mistake in regexp example in README [#87](https://github.com/lamoda/gonkey/pull/87) ([@JustSkiv](https://github.com/JustSkiv))

#### Authors: 4

- [@Alexey19](https://github.com/Alexey19)
- Alexey Tyuryumov (Alexey.Tyuryumov@acronis.com)
- Nikolay Oskin ([@NikolayOskin](https://github.com/NikolayOskin))
- Nikolay Tuzov ([@JustSkiv](https://github.com/JustSkiv))

---

# v1.4.0 (Fri Mar 19 2021)

#### 🚀 Enhancement

- fix: remove dependency from go-openapi [#82](https://github.com/lamoda/gonkey/pull/82) (Alexey.Tyuryumov@acronis.com [@Alexey19](https://github.com/Alexey19))

#### Authors: 2

- [@Alexey19](https://github.com/Alexey19)
- Alexey Tyuryumov (Alexey.Tyuryumov@acronis.com)

---

# v1.3.1 (Wed Mar 03 2021)

#### 🐛 Bug Fix

- Фикс проверки для #78 [#79](https://github.com/lamoda/gonkey/pull/79) ([@rsimkin](https://github.com/rsimkin))
- Automate release [#81](https://github.com/lamoda/gonkey/pull/81) (denis.fetinin@lamoda.ru [@What-If-I](https://github.com/What-If-I))

#### Authors: 3

- Denis ([@What-If-I](https://github.com/What-If-I))
- Denis Fetinin (denis.fetinin@lamoda.ru)
- Roman ([@rsimkin](https://github.com/rsimkin))
