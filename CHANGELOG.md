# Changelog

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
