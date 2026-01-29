# Changelog

## [1.7.3](https://github.com/entur/go-orchestrator/compare/v1.7.2...v1.7.3) (2026-01-29)


### Bug Fixes

* add contextId to logging entries ([#93](https://github.com/entur/go-orchestrator/issues/93)) ([671044b](https://github.com/entur/go-orchestrator/commit/671044beea6a07bc8fd7acd1571647c0b5b82cee))

## [1.7.2](https://github.com/entur/go-orchestrator/compare/v1.7.1...v1.7.2) (2025-11-10)


### Bug Fixes

* Miscellaneous (but important) fixes ([#78](https://github.com/entur/go-orchestrator/issues/78)) ([34d5332](https://github.com/entur/go-orchestrator/commit/34d5332207a33f83335619efaf1632fc00afc722))

## [1.7.1](https://github.com/entur/go-orchestrator/compare/v1.7.0...v1.7.1) (2025-09-11)


### Bug Fixes

* Inform user about invalid kinds, and suggest a list of possible alternatives ([#69](https://github.com/entur/go-orchestrator/issues/69)) ([21a4710](https://github.com/entur/go-orchestrator/commit/21a47107ac1270570b5bcb3b108e72f8c777138f))

## [1.7.0](https://github.com/entur/go-orchestrator/compare/v1.6.0...v1.7.0) (2025-09-10)


### Features

* miscellaneous changes for a better user experience ([#64](https://github.com/entur/go-orchestrator/issues/64)) ([9fcb760](https://github.com/entur/go-orchestrator/commit/9fcb7607ff466f502c7ffd7de16c8a662d278c20))

## [1.6.0](https://github.com/entur/go-orchestrator/compare/v1.5.1...v1.6.0) (2025-09-04)


### Features

* Added jsonschema validation tags to ManifestHeader ([#62](https://github.com/entur/go-orchestrator/issues/62)) ([7520677](https://github.com/entur/go-orchestrator/commit/7520677279c56d61469a2875dec9945a5a9fc387))

## [1.5.1](https://github.com/entur/go-orchestrator/compare/v1.5.0...v1.5.1) (2025-09-02)


### Bug Fixes

* Add support for []string and []Change in Result additions ([#60](https://github.com/entur/go-orchestrator/issues/60)) ([bec2349](https://github.com/entur/go-orchestrator/commit/bec2349d0c5ccbe2bb54ddcf2811425a7899ccdc))

## [1.5.0](https://github.com/entur/go-orchestrator/compare/v1.4.0...v1.5.0) (2025-09-02)


### Features

* Better Go-Orchestrator SDK experience ([#57](https://github.com/entur/go-orchestrator/issues/57)) ([110e09e](https://github.com/entur/go-orchestrator/commit/110e09e3320de8681dd69708a29a7bba5397d110))

## [1.4.0](https://github.com/entur/go-orchestrator/compare/v1.3.0...v1.4.0) (2025-08-11)


### Features

* Added support for new origin repository fields ([#51](https://github.com/entur/go-orchestrator/issues/51)) ([556aaab](https://github.com/entur/go-orchestrator/commit/556aaab443d92ee7d89e463354f9df1749a559d6))

## [1.3.0](https://github.com/entur/go-orchestrator/compare/v1.2.2...v1.3.0) (2025-06-04)


### Features

* Added IsDone function to Result ([#40](https://github.com/entur/go-orchestrator/issues/40)) ([e4bd7c1](https://github.com/entur/go-orchestrator/commit/e4bd7c1c6ea31e78bd79f3719708c349e1d1f544))

## [1.2.2](https://github.com/entur/go-orchestrator/compare/v1.2.1...v1.2.2) (2025-06-04)


### Bug Fixes

* Don't attempt unmarshalling non-json request responses ([#36](https://github.com/entur/go-orchestrator/issues/36)) ([0bf5858](https://github.com/entur/go-orchestrator/commit/0bf5858b0e7221b5c8866cd49c2d31f281a49c4f))

## [1.2.1](https://github.com/entur/go-orchestrator/compare/v1.2.0...v1.2.1) (2025-06-04)


### Bug Fixes

* Fixed some requests failing (GCPAppProjectIDS) due to EOF error ([#34](https://github.com/entur/go-orchestrator/issues/34)) ([5d92b74](https://github.com/entur/go-orchestrator/commit/5d92b74688d592dff267c590faae763ee09da994))

## [1.2.0](https://github.com/entur/go-orchestrator/compare/v1.1.0...v1.2.0) (2025-06-04)


### Features

* Support mocking request directly ([#32](https://github.com/entur/go-orchestrator/issues/32)) ([5394dd8](https://github.com/entur/go-orchestrator/commit/5394dd8eda11db026cc03fb56f07a8f98126a0a6))

## [1.1.0](https://github.com/entur/go-orchestrator/compare/v1.0.0...v1.1.0) (2025-06-03)


### Features

* Added ctx request cache ([#29](https://github.com/entur/go-orchestrator/issues/29)) ([3ca2833](https://github.com/entur/go-orchestrator/commit/3ca283379bbe35d7ec854885964c0cfacb3bdfc9))
* Added support for manifest handler middlewares too ([#31](https://github.com/entur/go-orchestrator/issues/31)) ([6566d29](https://github.com/entur/go-orchestrator/commit/6566d29c5565ebca6dae6cca440e7bfb490140ad))

## [1.0.0](https://github.com/entur/go-orchestrator/compare/v1.2.3...v1.0.0) (2025-05-27)


### ⚠ BREAKING CHANGES

* **refactor:** Rewrote Go Orchestrator to support multiple different ApiVersions and Kinds ([#23](https://github.com/entur/go-orchestrator/issues/23))
* Added iam lookup resource and tried to simplify SO creation a bit ([#9](https://github.com/entur/go-orchestrator/issues/9))

### Features

* Added iam lookup resource and tried to simplify SO creation a bit ([#9](https://github.com/entur/go-orchestrator/issues/9)) ([8e917e0](https://github.com/entur/go-orchestrator/commit/8e917e000ca7615db5399c8f2e4d6eef3a793969))
* Updated sdk to use common logging library ([b6bd03e](https://github.com/entur/go-orchestrator/commit/b6bd03e5f39df94d54ebd3032e115ba0108d566c))


### Bug Fixes

* Change Result String() return value based on internal state ([13562d2](https://github.com/entur/go-orchestrator/commit/13562d27325e1f47d7234bcb386802efa7d6ce63))
* even more logging for error on send ([a68be44](https://github.com/entur/go-orchestrator/commit/a68be44ecae0f8c0d1ef12dd0d17ec4f6c4818a6))
* Fixed test, added prefix to logged fields to ensure no duplicates, removed request_id again from MockEvent ([f298275](https://github.com/entur/go-orchestrator/commit/f298275f7cad015a42cd01afa1aedeca00490411))
* forgot about go ([34f0fb5](https://github.com/entur/go-orchestrator/commit/34f0fb548d08010ccc3a7e4226dd33dbf389dd58))
* less mutation, more logging ([dd86cd5](https://github.com/entur/go-orchestrator/commit/dd86cd526c50b913e5e96af0021865affffa902b))
* Lock mutex when iterating over topics ([76448cb](https://github.com/entur/go-orchestrator/commit/76448cbca844983ae2fcd63502a5f69efc1a815c))
* move logger init to within handler creation ([68257b1](https://github.com/entur/go-orchestrator/commit/68257b1c6700cb05886b47662d1d632d299a9880))
* options for mock event ([#16](https://github.com/entur/go-orchestrator/issues/16)) ([765a668](https://github.com/entur/go-orchestrator/commit/765a6687f51223ea3d080ddcb5c17521c3534a98))
* ready for first release ([#2](https://github.com/entur/go-orchestrator/issues/2)) ([249bbfc](https://github.com/entur/go-orchestrator/commit/249bbfc22a8dcacd26af5f6d8c6813ebcaa1d2b0))
* **refactor:** Miscellaneous changes relating to docs and function names ([#19](https://github.com/entur/go-orchestrator/issues/19)) ([3e9c7b0](https://github.com/entur/go-orchestrator/commit/3e9c7b0f41298cada237593d662e469eaefaf7c8))
* release ready ([7fa8c21](https://github.com/entur/go-orchestrator/commit/7fa8c214c30ae0e42b60648b23cf65d439edc4a8))
* Renamed logged fields in test output as well ([b509eb5](https://github.com/entur/go-orchestrator/commit/b509eb544b9dfa06d6319e00199bf89e4338b231))
* requestId not request_id ([#12](https://github.com/entur/go-orchestrator/issues/12)) ([58a6ee7](https://github.com/entur/go-orchestrator/commit/58a6ee73b32e5371e1a214ea0cdaafaf61ddb70b))
* structured result and accessible raw arrays for testing ([67c441b](https://github.com/entur/go-orchestrator/commit/67c441beb8a3085ef2a6cca55cf641c9fe85b276))
* test without timestamp for example testing ([bab4bdb](https://github.com/entur/go-orchestrator/commit/bab4bdb689ad01193fe0a2731dee13cb37d32b23))


### Miscellaneous Chores

* **refactor:** Rewrote Go Orchestrator to support multiple different ApiVersions and Kinds ([#23](https://github.com/entur/go-orchestrator/issues/23)) ([f771e54](https://github.com/entur/go-orchestrator/commit/f771e543ecc4b9fa8daa37c5779e7dbfe681aee8))
* release 1.0.0 ([0062970](https://github.com/entur/go-orchestrator/commit/00629702503175c4edb34b0b2d4f6204c428d7e9))

## [0.2.3](https://github.com/entur/go-orchestrator/compare/v1.2.2...v1.2.3) (2025-05-21)

### Bug Fixes

- **refactor:** Miscellaneous changes relating to docs and function names ([#19](https://github.com/entur/go-orchestrator/issues/19)) ([3e9c7b0](https://github.com/entur/go-orchestrator/commit/3e9c7b0f41298cada237593d662e469eaefaf7c8))

## [0.2.2](https://github.com/entur/go-orchestrator/compare/v1.2.1...v1.2.2) (2025-05-20)

### Bug Fixes

- options for mock event ([#16](https://github.com/entur/go-orchestrator/issues/16)) ([765a668](https://github.com/entur/go-orchestrator/commit/765a6687f51223ea3d080ddcb5c17521c3534a98))

## [0.2.1](https://github.com/entur/go-orchestrator/compare/v1.2.0...v1.2.1) (2025-05-19)

### Bug Fixes

- requestId not request_id ([#12](https://github.com/entur/go-orchestrator/issues/12)) ([58a6ee7](https://github.com/entur/go-orchestrator/commit/58a6ee73b32e5371e1a214ea0cdaafaf61ddb70b))

## [0.2.0](https://github.com/entur/go-orchestrator/compare/v2.0.0...v1.2.0) (2025-05-16)

### Features

- Added iam lookup resource and tried to simplify SO creation a bit ([#9](https://github.com/entur/go-orchestrator/issues/9))
- Added iam lookup resource and tried to simplify SO creation a bit ([#9](https://github.com/entur/go-orchestrator/issues/9)) ([8e917e0](https://github.com/entur/go-orchestrator/commit/8e917e000ca7615db5399c8f2e4d6eef3a793969))
- Updated sdk to use common logging library ([b6bd03e](https://github.com/entur/go-orchestrator/commit/b6bd03e5f39df94d54ebd3032e115ba0108d566c))

### Bug Fixes

- Change Result String() return value based on internal state ([13562d2](https://github.com/entur/go-orchestrator/commit/13562d27325e1f47d7234bcb386802efa7d6ce63))
- even more logging for error on send ([a68be44](https://github.com/entur/go-orchestrator/commit/a68be44ecae0f8c0d1ef12dd0d17ec4f6c4818a6))
- Fixed test, added prefix to logged fields to ensure no duplicates, removed request_id again from MockEvent ([f298275](https://github.com/entur/go-orchestrator/commit/f298275f7cad015a42cd01afa1aedeca00490411))
- forgot about go ([34f0fb5](https://github.com/entur/go-orchestrator/commit/34f0fb548d08010ccc3a7e4226dd33dbf389dd58))
- less mutation, more logging ([dd86cd5](https://github.com/entur/go-orchestrator/commit/dd86cd526c50b913e5e96af0021865affffa902b))
- Lock mutex when iterating over topics ([76448cb](https://github.com/entur/go-orchestrator/commit/76448cbca844983ae2fcd63502a5f69efc1a815c))
- move logger init to within handler creation ([68257b1](https://github.com/entur/go-orchestrator/commit/68257b1c6700cb05886b47662d1d632d299a9880))
- ready for first release ([#2](https://github.com/entur/go-orchestrator/issues/2)) ([249bbfc](https://github.com/entur/go-orchestrator/commit/249bbfc22a8dcacd26af5f6d8c6813ebcaa1d2b0))
- release ready ([7fa8c21](https://github.com/entur/go-orchestrator/commit/7fa8c214c30ae0e42b60648b23cf65d439edc4a8))
- Renamed logged fields in test output as well ([b509eb5](https://github.com/entur/go-orchestrator/commit/b509eb544b9dfa06d6319e00199bf89e4338b231))
- structured result and accessible raw arrays for testing ([67c441b](https://github.com/entur/go-orchestrator/commit/67c441beb8a3085ef2a6cca55cf641c9fe85b276))
- test without timestamp for example testing ([bab4bdb](https://github.com/entur/go-orchestrator/commit/bab4bdb689ad01193fe0a2731dee13cb37d32b23))

### Miscellaneous Chores

- release 1.2.0 ([a637238](https://github.com/entur/go-orchestrator/commit/a6372385be9d8a61ff2045f75b24b03b0f737863))

## [0.1.0](https://github.com/entur/go-orchestrator/compare/v1.0.0...v1.1.0) (2025-05-14)

### Features

- Updated sdk to use common logging library ([b6bd03e](https://github.com/entur/go-orchestrator/commit/b6bd03e5f39df94d54ebd3032e115ba0108d566c))

## 1.0.0 (2025-05-14)

### Features

- Updated sdk to use common logging library ([b6bd03e](https://github.com/entur/go-orchestrator/commit/b6bd03e5f39df94d54ebd3032e115ba0108d566c))

### Bug Fixes

- Change Result String() return value based on internal state ([13562d2](https://github.com/entur/go-orchestrator/commit/13562d27325e1f47d7234bcb386802efa7d6ce63))
- even more logging for error on send ([a68be44](https://github.com/entur/go-orchestrator/commit/a68be44ecae0f8c0d1ef12dd0d17ec4f6c4818a6))
- Fixed test, added prefix to logged fields to ensure no duplicates, removed request_id again from MockEvent ([f298275](https://github.com/entur/go-orchestrator/commit/f298275f7cad015a42cd01afa1aedeca00490411))
- forgot about go ([34f0fb5](https://github.com/entur/go-orchestrator/commit/34f0fb548d08010ccc3a7e4226dd33dbf389dd58))
- less mutation, more logging ([dd86cd5](https://github.com/entur/go-orchestrator/commit/dd86cd526c50b913e5e96af0021865affffa902b))
- Lock mutex when iterating over topics ([76448cb](https://github.com/entur/go-orchestrator/commit/76448cbca844983ae2fcd63502a5f69efc1a815c))
- move logger init to within handler creation ([68257b1](https://github.com/entur/go-orchestrator/commit/68257b1c6700cb05886b47662d1d632d299a9880))
- ready for first release ([#2](https://github.com/entur/go-orchestrator/issues/2)) ([249bbfc](https://github.com/entur/go-orchestrator/commit/249bbfc22a8dcacd26af5f6d8c6813ebcaa1d2b0))
- Renamed logged fields in test output as well ([b509eb5](https://github.com/entur/go-orchestrator/commit/b509eb544b9dfa06d6319e00199bf89e4338b231))
- structured result and accessible raw arrays for testing ([67c441b](https://github.com/entur/go-orchestrator/commit/67c441beb8a3085ef2a6cca55cf641c9fe85b276))
- test without timestamp for example testing ([bab4bdb](https://github.com/entur/go-orchestrator/commit/bab4bdb689ad01193fe0a2731dee13cb37d32b23))
