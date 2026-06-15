# Changelog

## [0.22.0](https://github.com/adaouat/bifrost/compare/v0.21.0..v0.22.0) - 2026-06-15

### 🚀 Features

- *(config)* Reject unsupported strategy values - ([35709f3](https://github.com/adaouat/bifrost/commit/35709f3dd95df8197afe47dbc70fadcd9ebd2ce5)) by @bchatard

- *(config)* Reject empty hook cmd at load time - ([bb39dbc](https://github.com/adaouat/bifrost/commit/bb39dbc7a67ce7c0b1e482c5f06f48a3794e9259)) by @bchatard

- *(deploy)* Cancel the deploy on SIGINT via cmd.Context() - ([c897858](https://github.com/adaouat/bifrost/commit/c897858a8d73c576a84db9c2d0ca2662d16b8d3f)) by @bchatard

- *(deps)* Bump forge to v0.17.0 - ([2f285a8](https://github.com/adaouat/bifrost/commit/2f285a8df58178ef0a8e61afa8196659e46eca28)) by @bchatard


### 🐛 Bug Fixes

- *(deploy)* Log purge-plan errors instead of discarding them - ([e26a80f](https://github.com/adaouat/bifrost/commit/e26a80fc36ce206542f333602fa70142eb7a146f)) by @bchatard

- *(transport)* Bound agent downloads with an http.Client timeout - ([5832a68](https://github.com/adaouat/bifrost/commit/5832a68b9d4893251b7b9605eb3b7458230fc6eb)) by @bchatard

- *(tui)* Raise event-stream line limit above 64 KB - ([7bad1af](https://github.com/adaouat/bifrost/commit/7bad1afc9870bfa5734788ece765256c4d85d33d)) by @bchatard


### 🚜 Refactor

- *(cmd)* Extract shared config-resolution and hook helpers - ([59ba701](https://github.com/adaouat/bifrost/commit/59ba701d4b99df31c932eb0503857d9c03572883)) by @bchatard

- *(deploy)* Collapse the 8 hook-stage blocks into one helper - ([eee16f7](https://github.com/adaouat/bifrost/commit/eee16f7d56e154ca4ce00c21bae951bc2d6bb5f6)) by @bchatard

- *(hooks)* Use errors.AsType to match ssh.go - ([0d599a7](https://github.com/adaouat/bifrost/commit/0d599a724ad3bb2337e4aa3273d6fcef31893056)) by @bchatard

- *(transport)* Chmod the agent via SFTP, not a remote shell - ([2c64e60](https://github.com/adaouat/bifrost/commit/2c64e60ce40edf3805a03c34eacfb8df7f499eb6)) by @bchatard


### 📚 Documentation

- *(testing)* Allow t.TempDir in deployer FS-logic tests - ([f476cd6](https://github.com/adaouat/bifrost/commit/f476cd6e40adea7fe087fefcc17445f1482deadd)) by @bchatard


### 🧪 Testing

- *(cmd)* Assert hook output is visible over SSH (S11) - ([2789e1d](https://github.com/adaouat/bifrost/commit/2789e1da4bc3cca8c8934895adae593fecf74dea)) by @bchatard

- *(cmd)* Confirm deploy error via stderr, not the exec exit code - ([236e130](https://github.com/adaouat/bifrost/commit/236e130b708c9a9527034dd430ccce6f3ae69b24)) by @bchatard

- *(testutil)* Unique known_hosts path to fix CI host-key race - ([bca2aa2](https://github.com/adaouat/bifrost/commit/bca2aa26182846018807fdb6202fe8491c868cc1)) by @bchatard

## [0.21.0](https://github.com/adaouat/bifrost/compare/v0.20.0..v0.21.0) - 2026-06-14

### 🚀 Features

- *(transport)* Verify agent binary checksum before caching - ([c332ddb](https://github.com/adaouat/bifrost/commit/c332ddbab3ff802f76149189b2a50f66dcdc3153)) by @bchatard


### 🐛 Bug Fixes

- *(deploy)* Emit hook output as JSON events in --output json mode - ([a3675e2](https://github.com/adaouat/bifrost/commit/a3675e23e0aff60c81429717ce335f00eda41856)) by @bchatard

- *(deploy)* Classify hook and deploy runtime failures as exit 3 - ([11755f3](https://github.com/adaouat/bifrost/commit/11755f35fa4ff33560db7017349b1a5980670fbc)) by @bchatard

- *(transport)* Download agent as a raw binary, not a tar.gz - ([ec289ad](https://github.com/adaouat/bifrost/commit/ec289ad2df88593ad9ca7b985da0e9ef4f102559)) by @bchatard

- *(transport)* Shell-quote interpolated values in remote commands - ([689e901](https://github.com/adaouat/bifrost/commit/689e9018cdcc3145bc46ce11df65fa2cf646c073)) by @bchatard


### 📚 Documentation

- *(tasks)* Add M19 hardening and M20 cleanup milestones - ([bdd04a4](https://github.com/adaouat/bifrost/commit/bdd04a47b43b27921a1bda35ed3ea7bb49a344ee)) by @bchatard

- *(tasks)* Record the container-readiness fix under M19 - ([03b248c](https://github.com/adaouat/bifrost/commit/03b248cfec216aa44845dfa04c069ab3f9ba3caa)) by @bchatard

- *(testing)* Require container readiness to gate on a setup marker - ([639c94d](https://github.com/adaouat/bifrost/commit/639c94d8b1bdb2524c2d71dc19bf68b95c390eb2)) by @bchatard

- Reconcile specs and rules with shipped behaviour - ([1ed87d6](https://github.com/adaouat/bifrost/commit/1ed87d6ef4813b3b3f5194a6ac2838b438c73882)) by @bchatard

- Move M18 working-doc artifacts to .claude/plans - ([08d7b6e](https://github.com/adaouat/bifrost/commit/08d7b6e3788a306c9de76fa9fefbb8af7ff959a3)) by @bchatard


### 🧪 Testing

- *(testutil)* Wait for container setup to finish, not just the port - ([e837a96](https://github.com/adaouat/bifrost/commit/e837a966817d5cc032d7f3856a8569e3928e3282)) by @bchatard


### ⚙️ Miscellaneous Tasks

- Run the integration test suite in CI - ([98d8254](https://github.com/adaouat/bifrost/commit/98d82543dc7ad7a1418973ec0414d5b777079740)) by @bchatard

## [0.20.0](https://github.com/adaouat/bifrost/compare/v0.19.1..v0.20.0) - 2026-06-13

### 🚀 Features

- *(config)* Update config init scaffold for 8-hook lifecycle - ([2b7260e](https://github.com/adaouat/bifrost/commit/2b7260eed0922a38ee4be88430ea1fa48723bd9d)) by @bchatard

- *(hooks)* Expand hook lifecycle to 8 symmetric pre/post stages - ([3f298ae](https://github.com/adaouat/bifrost/commit/3f298ae3db8952436075fcb43e7835c92e797c80)) by @bchatard


### 🐛 Bug Fixes

- *(heraut)* Update config to add new required field - ([ede8b33](https://github.com/adaouat/bifrost/commit/ede8b33b7c193bb5772a129f89ae0ae74a1c2c3c)) by @bchatard


### 🚜 Refactor

- *(config)* Rename ServerConfig to Server - ([9357cef](https://github.com/adaouat/bifrost/commit/9357cefc868cd678431288cc1157b8f77414f13d)) by @bchatard


### 📚 Documentation

- *(adr)* Record hook lifecycle granularity decision (ADR-0012) - ([1b9faf7](https://github.com/adaouat/bifrost/commit/1b9faf70a2e9d6cc328c6393acf24732aab1fbb4)) by @bchatard

- *(claude)* Document transport package and fix lint command name - ([bdc57e4](https://github.com/adaouat/bifrost/commit/bdc57e4b11ccbfae29b4a58d48f064d4af830b96)) by @bchatard

- *(config)* Add annotated sample .bifrost.yml - ([3887416](https://github.com/adaouat/bifrost/commit/3887416b913253f7af1fe548ff479ef02f1eaea4)) by @bchatard

- *(plans)* Record hook lifecycle granularity implementation plan - ([fd32c1e](https://github.com/adaouat/bifrost/commit/fd32c1eaecfa6abf576a4d19480f4b3aaeacaccc)) by @bchatard

- *(schema)* Update bifrost.schema.json for 8-hook lifecycle - ([685d2fe](https://github.com/adaouat/bifrost/commit/685d2fe798d3f2f9684d04ebb94175c953472829)) by @bchatard

- *(specs)* Add hook lifecycle redesign spec - ([fa6df86](https://github.com/adaouat/bifrost/commit/fa6df86df816b449285e319de4705660d8451255)) by @bchatard

- *(specs)* Update spec 02 for 8-hook lifecycle - ([19f21ed](https://github.com/adaouat/bifrost/commit/19f21ed29317848b029c3618f186b031ac7e2b76)) by @bchatard

- *(specs)* Rewrite spec 05 hook lifecycle table for 8 hooks - ([56f0d69](https://github.com/adaouat/bifrost/commit/56f0d6979dd3a52b6f4eff630b4559d069b94fac)) by @bchatard

- *(specs)* Update specs 03 and 06 for 8-hook lifecycle - ([9d56f37](https://github.com/adaouat/bifrost/commit/9d56f373d7c1887cb8720eea9357f45428d66f71)) by @bchatard

- *(tasks)* Mark M18 hook lifecycle granularity complete - ([5619cf4](https://github.com/adaouat/bifrost/commit/5619cf472f2e89eb17253eed53e0a0105af4f19c)) by @bchatard

- Update bifrost.sample.yml for 8-hook lifecycle - ([30786d6](https://github.com/adaouat/bifrost/commit/30786d65827c67ccf1e7cd34b6b25801b9cbd809)) by @bchatard


### 🧪 Testing

- *(cmd)* Cover all 8 hook lifecycle points and verify deploy ordering - ([1cc766c](https://github.com/adaouat/bifrost/commit/1cc766c079e27480ee5e56dce5b9d16cb7c5ef19)) by @bchatard

- *(config)* Assert old hook key names are rejected by strict YAML - ([6711c1d](https://github.com/adaouat/bifrost/commit/6711c1dee0790108e32fd6dd939e27450b64cb76)) by @bchatard

## [0.19.1](https://github.com/adaouat/bifrost/compare/v0.19.0..v0.19.1) - 2026-06-12

### 🐛 Bug Fixes

- *(deps)* Bump forge to v0.16.0 - ([15c818c](https://github.com/adaouat/bifrost/commit/15c818c838a3ae9b89dcdab814c58d3790a7bdc6)) by @bchatard

## [0.19.0](https://github.com/adaouat/bifrost/compare/v0.18.0..v0.19.0) - 2026-06-11

### 🚀 Features

- *(config)* Add JSON Schema and wire it into config init - ([c96bc6f](https://github.com/adaouat/bifrost/commit/c96bc6fa96570ec6b227c8a852e3730111194039)) by @bchatard


### 📚 Documentation

- *(roadmap)* Mark M17 server-resolution and staging-path tests done - ([2baff8b](https://github.com/adaouat/bifrost/commit/2baff8b4b4b7178475cd0c799cc86a154cbdcb67)) by @bchatard

- *(specs)* Reconcile specs 01-03,07 with shipped implementation - ([cc7ecaa](https://github.com/adaouat/bifrost/commit/cc7ecaad009cde875d8f27a4f0104ebed3e8b676)) by @bchatard


### 🧪 Testing

- *(cmd)* Integration test for --agent-binary download bypass - ([eee82d5](https://github.com/adaouat/bifrost/commit/eee82d5276f6ee0f9b49f45970617eb6de3b8100)) by @bchatard

- *(cmd)* Integration test for SSH auth failure exit code - ([dba401c](https://github.com/adaouat/bifrost/commit/dba401cc3f8b0b202a754cbbd9394df9258531ac)) by @bchatard

- *(cmd)* Integration test for unknown remote arch exit code - ([9d31be2](https://github.com/adaouat/bifrost/commit/9d31be2d83695e549f8eec8ea14e93d00e35dad4)) by @bchatard

- *(cmd)* Integration test for staging cleanup on agent failure - ([a7101b8](https://github.com/adaouat/bifrost/commit/a7101b858e17a0b54f4746097629764e0003a1fa)) by @bchatard

- *(config)* Cover flat config generator merge precedence - ([d92beac](https://github.com/adaouat/bifrost/commit/d92beac2bd6f758e44a144b17885d83d726abd1d)) by @bchatard

## [0.18.0](https://github.com/adaouat/bifrost/compare/v0.17.0..v0.18.0) - 2026-06-10

### 🚀 Features

- *(whatsnew)* Embed CHANGELOG.md as the offline fallback - ([8893019](https://github.com/adaouat/bifrost/commit/8893019e4d0c5010fa174c62dd5aea5adbf38136)) by @bchatard


### 💼 Other

- *(deps)* Bump forge to v0.15.0 - ([8eeded6](https://github.com/adaouat/bifrost/commit/8eeded6a3f532c0937177e2c041c329283dd8537)) by @bchatard

## [0.17.0](https://github.com/adaouat/bifrost/compare/v0.16.0..v0.17.0) - 2026-06-10

### 🚀 Features

- *(cmd)* Add whatsnew command, adopt forge's shared hint wiring - ([3ddd3bf](https://github.com/adaouat/bifrost/commit/3ddd3bf77857c677ce446f7d21b717baee700a39)) by @bchatard


### 💼 Other

- *(deps)* Bump forge to v0.14.0 - ([8e10aa1](https://github.com/adaouat/bifrost/commit/8e10aa12e94260b9568d498f8073b71527620786)) by @bchatard

## [0.16.0](https://github.com/adaouat/bifrost/compare/v0.15.0..v0.16.0) - 2026-06-09

### 🚀 Features

- *(cmd)* M16 — release list over SSH - ([ebe03f0](https://github.com/adaouat/bifrost/commit/ebe03f00eed34ed7a390993be17f32c0243774d2)) by @bchatard

- *(cmd)* M16 — release activate over SSH - ([506fc37](https://github.com/adaouat/bifrost/commit/506fc373cebf56dfb2ccf85fe86a0f8d71607cd2)) by @bchatard

- *(cmd)* M16 — release rollback over SSH - ([10beaf1](https://github.com/adaouat/bifrost/commit/10beaf1c9b56817226680447a88b694444ad0366)) by @bchatard

- *(deploy)* Add operator-debugging diagnostics to the atomic deploy path - ([d3d4c52](https://github.com/adaouat/bifrost/commit/d3d4c52a5919045597ea347f56efd169b14a439b)) by @bchatard


### 💼 Other

- *(deps)* Bump forge to v0.11.0 - ([c32844d](https://github.com/adaouat/bifrost/commit/c32844d8b5d36d02275d34cc5f19e9a81e293192)) by @bchatard

- *(deps)* Bump forge to v0.11.1 - ([0d50b76](https://github.com/adaouat/bifrost/commit/0d50b76a86a8ceea539d6ff5e4c41834a36a4d4e)) by @bchatard


### 🧪 Testing

- *(cmd)* M16 — integration tests for release commands over SSH - ([32a7774](https://github.com/adaouat/bifrost/commit/32a7774af67da38c8eff0248886f52c2a35bc29e)) by @bchatard

## [0.15.0](https://github.com/adaouat/bifrost/compare/v0.14.0..v0.15.0) - 2026-06-08

### 🚀 Features

- *(cmd)* M15 — verify sequential multi-server deploy loop - ([b43f2ba](https://github.com/adaouat/bifrost/commit/b43f2babd2ba4110403b57dd20066c12eb40256e)) by @bchatard

- *(tui)* Per-server header line for multi-server deploy output - ([1048acc](https://github.com/adaouat/bifrost/commit/1048acc332e0b0c499281350e9d34b433678438f)) by @bchatard


### 🐛 Bug Fixes

- *(cmd)* Copy artifact in dry-run sudo hook integration test - ([188cfb6](https://github.com/adaouat/bifrost/commit/188cfb68083e6ece3a4a59561e1cdc2a7fdb573f)) by @bchatard


### 🧪 Testing

- *(cmd)* M15 — two-container sequential multi-server deploy E2E - ([fe2a352](https://github.com/adaouat/bifrost/commit/fe2a352ea7a2d5bc932abaa0b1b8062380690e7a)) by @bchatard

## [0.14.0](https://github.com/adaouat/bifrost/compare/v0.13.0..v0.14.0) - 2026-06-07

### 🚀 Features

- *(cmd)* Implement SSH client mode deploy - ([65e37c3](https://github.com/adaouat/bifrost/commit/65e37c3fd714fa843c530b3b97392a7dacc7e2bb)) by @bchatard

- *(cmd)* M14 — integration test for SSH client mode deploy - ([2c3c6d7](https://github.com/adaouat/bifrost/commit/2c3c6d781a12b48a299ea96f77a4af5321055179)) by @bchatard

- *(config)* Add flat config generator for SSH client mode - ([660adf2](https://github.com/adaouat/bifrost/commit/660adf29c4bdc12d028f7e39638b4874034b3cc6)) by @bchatard


### 🐛 Bug Fixes

- Readable --help usage block (bump forge to v0.9.0) - ([674f9ef](https://github.com/adaouat/bifrost/commit/674f9eff6c99d3e1b523908e8a93c402e19cfdcd)) by @bchatard

## [0.13.0](https://github.com/adaouat/bifrost/compare/v0.12.0..v0.13.0) - 2026-06-05

### 🚀 Features

- *(cmd)* Adopt forge cli.Run + theme (drop direct fang) - ([de721ec](https://github.com/adaouat/bifrost/commit/de721ec7e6f7f17c1a77ee2bf2a2aa85e865b1a5)) by @bchatard

- *(cmd)* Brand huh prompts with the Aurora theme - ([0669ca5](https://github.com/adaouat/bifrost/commit/0669ca5e85c080d0b02bee83ae27ef97ab2e97b8)) by @bchatard

## [0.12.0](https://github.com/adaouat/bifrost/compare/v0.11.0..v0.12.0) - 2026-06-05

### 🚀 Features

- *(cmd)* Apply the Aurora fang theme (forge v0.7.0) - ([9f88bff](https://github.com/adaouat/bifrost/commit/9f88bff9559ec0ac5a1831fd2c8737f8395d4120)) by @bchatard


### 💼 Other

- *(deps)* Align cobra to 1.10.2 - ([6288d98](https://github.com/adaouat/bifrost/commit/6288d987a69fc7b4501193918eaf077b189bb4b9)) by @bchatard


### ⚙️ Miscellaneous Tasks

- *(release)* Use forge's release-setup composite action - ([973d533](https://github.com/adaouat/bifrost/commit/973d5332428912163e3f250bfbf45dc35810beab)) by @bchatard

## [0.11.0](https://github.com/adaouat/bifrost/compare/v0.10.0..v0.11.0) - 2026-06-04

### 🚀 Features

- *(tui)* Number deploy steps via the shared spinner - ([3d12e7b](https://github.com/adaouat/bifrost/commit/3d12e7bfa8a6a3404660f39fd5ec3a89c62ca5d4)) by @bchatard


### 🐛 Bug Fixes

- *(hooks)* Don't echo stderr twice in allow_fail warning - ([29f8dfd](https://github.com/adaouat/bifrost/commit/29f8dfdebb95385b7e7d854aade43885c7b6b33a)) by @bchatard


### 💼 Other

- *(deps)* Bump golang.org/x/crypto to v0.52.0 - ([1ea0972](https://github.com/adaouat/bifrost/commit/1ea097282af66c32e001cec8de2bf5960c6bef4c)) by @bchatard

- *(mise)* Align goreleaser pin to 2.16 - ([dc118e7](https://github.com/adaouat/bifrost/commit/dc118e712750d7735c61139ab37a4810ac88f88a)) by @bchatard

- Bump Go toolchain lock to 1.26.4 - ([779d767](https://github.com/adaouat/bifrost/commit/779d7674308492168bf800f8c298581144db99f2)) by @bchatard

- Depend on published forge v0.6.2 (drop replace) - ([1ee04b4](https://github.com/adaouat/bifrost/commit/1ee04b46cd21ec755f928ae1dc90430168e21980)) by @bchatard


### 🚜 Refactor

- *(cmd)* Wire forge updatecheck hint and version injection - ([97613d9](https://github.com/adaouat/bifrost/commit/97613d91b252b904591b109130eed8953d71eb5a)) by @bchatard

- *(cmderr)* Adopt forge exitcode - ([6a8e2d5](https://github.com/adaouat/bifrost/commit/6a8e2d51db87a1c924a452dbf255e9d1075e7572)) by @bchatard

- *(cmderr)* Use named exit codes from forge - ([b04ab9c](https://github.com/adaouat/bifrost/commit/b04ab9c8de9c436016dcccf645c17f6003d3abcb)) by @bchatard

- *(cmdutil)* Resolve config path via forge.Resolver - ([58bc083](https://github.com/adaouat/bifrost/commit/58bc083fddf45d4b1f5fabc0f7e4d1c298149069)) by @bchatard

- *(config)* Use forge loader and ValidationError - ([38feebb](https://github.com/adaouat/bifrost/commit/38feebb8b1385d69b5e6fc7d9fde6a2395a64619)) by @bchatard

- *(config)* Emit forge's config error wording - ([0d880b1](https://github.com/adaouat/bifrost/commit/0d880b1fc9e4af009a8dbb67bd1f9fec5caec518)) by @bchatard

- *(hooks)* Run hooks through forge exec.Runner - ([bd36ca8](https://github.com/adaouat/bifrost/commit/bd36ca8193f737e3004324d4d542296a1814411c)) by @bchatard

- *(tui)* Adopt forge ui and de-globalize output mode - ([961035e](https://github.com/adaouat/bifrost/commit/961035e72c0637283c4e3c859adf678e77f9d186)) by @bchatard

- *(tui)* Run purge and step lines via forge.Spinner - ([4a3b68f](https://github.com/adaouat/bifrost/commit/4a3b68fd8ba79bc586adbe1dfd1cb2171457ec07)) by @bchatard

- *(tui)* Drop unused style vars - ([e3744af](https://github.com/adaouat/bifrost/commit/e3744af26639e43166562e882a86acfabdb3b05f)) by @bchatard


### 📚 Documentation

- *(readme)* Add install instructions - ([c67e916](https://github.com/adaouat/bifrost/commit/c67e9164fc441180cad8534645da06b5a4db7ade)) by @bchatard


### ⚙️ Miscellaneous Tasks

- *(release)* Inject build version via ldflags - ([29454a6](https://github.com/adaouat/bifrost/commit/29454a6eeba5eb22e02fd76c6ff147c89c22cd24)) by @bchatard

- *(release)* Converge to raw-binary asset model - ([bcc0886](https://github.com/adaouat/bifrost/commit/bcc0886b5a30b3cd6f2ecda8778514ae1dfeb4e8)) by @bchatard

- *(release)* Release bifrost via heraut (build-only goreleaser) - ([ee92c11](https://github.com/adaouat/bifrost/commit/ee92c11592c8f9d2fdf4112144d49e128700e877)) by @bchatard

- Gitignore goreleaser dist output - ([a34b153](https://github.com/adaouat/bifrost/commit/a34b153fbf514e85f32b7ba84c951f686a529481)) by @bchatard

- Call forge's shared lint/test reusable workflow - ([0a3883a](https://github.com/adaouat/bifrost/commit/0a3883a47737d312ba69ca9bd37c366238cf3a08)) by @bchatard

- Re-pin forge go-ci to v0.6.1 (required coverage-threshold) - ([48e3342](https://github.com/adaouat/bifrost/commit/48e33423e8619747d4ff47011239c6c5ebc5f512)) by @bchatard

- Publish Homebrew cask to adaouat/homebrew-tap - ([9eab935](https://github.com/adaouat/bifrost/commit/9eab93554984aa60082aa6df8fded4a256718b1f)) by @bchatard

## [0.10.0](https://github.com/adaouat/bifrost/compare/v0.9.0..v0.10.0) - 2026-06-02

### 🚀 Features

- *(cmd)* Add --agent-binary flag to deploy and SSH-capable release commands - ([98fc9c9](https://github.com/adaouat/bifrost/commit/98fc9c9109f75d4fc9acfa9aac7c0838c696e47a)) by @bchatard

- *(transport)* Add SSH client wrapper with auth chain and strict host keys - ([4dfeb0b](https://github.com/adaouat/bifrost/commit/4dfeb0b8a5c0b3cb53d625095baebda2016b0c1a)) by @bchatard

- *(transport)* Add SFTP wrapper for upload, mkdir, chmod - ([6f98610](https://github.com/adaouat/bifrost/commit/6f986103880cf71437647b4697513fd29a44990c)) by @bchatard

- *(transport)* Add remote staging directory lifecycle - ([6e245aa](https://github.com/adaouat/bifrost/commit/6e245aabfbc2c7c4f82114eec9c7445253b2a238)) by @bchatard

- *(transport)* Add agent binary resolution with arch detection and cache - ([75d94f2](https://github.com/adaouat/bifrost/commit/75d94f22beb7e9c6e39e4d3a9201ae108bece73c)) by @bchatard

## [0.9.0](https://github.com/adaouat/bifrost/compare/v0.8.1..v0.9.0) - 2026-06-02

### 🚀 Features

- *(config)* Add server config schema, validation, and merge resolution (M12) - ([81d729d](https://github.com/adaouat/bifrost/commit/81d729dc1bb39df15fa6728106bffb17f07b48d9)) by @bchatard


### 📚 Documentation

- *(v1)* Add JSON Schema task to M17 - ([56297c7](https://github.com/adaouat/bifrost/commit/56297c75b8341a1856d62a58965981644b2550cf)) by @bchatard

## [0.8.1](https://github.com/adaouat/bifrost/compare/v0.8.0..v0.8.1) - 2026-06-01

### 🚜 Refactor

- *(strategy)* Introduce Deployer interface and move atomic deploy logic - ([08f918b](https://github.com/adaouat/bifrost/commit/08f918b3389da4e8e917e0be58884ed05b2e24d1)) by @bchatard


### 📚 Documentation

- *(v1)* Add roadmap, specs, and ADRs for SSH orchestration + agent model - ([2e8f923](https://github.com/adaouat/bifrost/commit/2e8f9233e37ab0325c3a0da65458213dda4201d4)) by @bchatard

## [0.8.0](https://github.com/adaouat/bifrost/compare/v0.7.1..v0.8.0) - 2026-05-31

### 🚀 Features

- *(cmd/deploy)* Emit JSON error event on step failure - ([9243eb7](https://github.com/adaouat/bifrost/commit/9243eb72247b0361040f16945117ffaab6be1185)) by @bchatard

- *(cmd/deploy)* Start/done JSON events for link, current_symlink, purge - ([934dc89](https://github.com/adaouat/bifrost/commit/934dc89a9c4f44f29c2da7238f8dbcd8747b7c1a)) by @bchatard

- *(cmd/deploy)* Dry-run Would purge line with actual release names - ([bcc62c6](https://github.com/adaouat/bifrost/commit/bcc62c6fb5c043ee7c52805a24432cbab58ef1bc)) by @bchatard

- *(cmd/deploy)* Show (sudo) suffix for sudo hooks in dry-run output - ([1d69311](https://github.com/adaouat/bifrost/commit/1d693118ad5740e5192564099502568ad025ba51)) by @bchatard

- *(cmd/deploy)* Spinner while purging old releases - ([d212288](https://github.com/adaouat/bifrost/commit/d212288285c518e2422aa5a2dd9727172342486e)) by @bchatard

- *(release/list)* Header line and ← current suffix in human output - ([94f7a06](https://github.com/adaouat/bifrost/commit/94f7a0606c1b23b3adb00a0ed640f0e74462751c)) by @bchatard

- *(tui)* Deploy header panel with bordered box - ([d4b5404](https://github.com/adaouat/bifrost/commit/d4b5404e943ea5a4817a8b0fa2f96410fa9057bd)) by @bchatard

- *(tui)* Per-step checkmark lines for deploy human output - ([1ab7ca6](https://github.com/adaouat/bifrost/commit/1ab7ca6aa2506cd3020a47ffeaef451a904335be)) by @bchatard

- *(tui)* Final summary line for deploy human output - ([5429429](https://github.com/adaouat/bifrost/commit/54294296ae148fd2fbfcd2898a2512695f70e3d9)) by @bchatard

- *(tui)* Step detail sub-lines and enriched JSON events - ([71cf30a](https://github.com/adaouat/bifrost/commit/71cf30ad29181bc136c5d6faa4f64847f79ff409)) by @bchatard

- *(tui)* Add title prefix to extraction progress bar - ([e381a84](https://github.com/adaouat/bifrost/commit/e381a843d18219224d5892de8abe2bd76fc8ba54)) by @bchatard


### 🐛 Bug Fixes

- *(strategy/atomic)* Track compressed bytes for extraction progress - ([58de9e2](https://github.com/adaouat/bifrost/commit/58de9e2f17b0631520c337f9585efe461a8cd93b)) by @bchatard

- *(strategy/atomic)* Protect active release from purge - ([3386860](https://github.com/adaouat/bifrost/commit/33868600c4d733930e233ad257f850a385e84932)) by @bchatard

- Address important review findings - ([4143c38](https://github.com/adaouat/bifrost/commit/4143c3878a76c5e762c3924d5b62318ab9440a52)) by @bchatard


### 🚜 Refactor

- Address suggestion review findings - ([8873db2](https://github.com/adaouat/bifrost/commit/8873db2c665cf84ac64821ec11ead5248a41c1f5)) by @bchatard


### 📚 Documentation

- *(claude)* Sync CLAUDE.md project layout with current codebase - ([95b9018](https://github.com/adaouat/bifrost/commit/95b9018f6711737be17ecfb58515ccd053a34a66)) by @bchatard

- *(specs)* Sync 06-tui-ux and 03-commands with current implementation - ([5fa71ef](https://github.com/adaouat/bifrost/commit/5fa71efefcf614bf173d6c10e04b906474289018)) by @bchatard

## [0.7.1](https://github.com/adaouat/bifrost/compare/v0.7.0..v0.7.1) - 2026-05-31

### 🧪 Testing

- *(cmd/deploy)* Integration test for release purge - ([7ccb14a](https://github.com/adaouat/bifrost/commit/7ccb14a410358dd0123fa6163df093f822a8bfc7)) by @bchatard

- *(cmd/release)* Integration test for activating non-existent release - ([b80ace3](https://github.com/adaouat/bifrost/commit/b80ace33d6186547d26397f607583a481620af96)) by @bchatard

- *(config)* Unit tests for merge logic - ([35b07a3](https://github.com/adaouat/bifrost/commit/35b07a301c2430bdc48ad8153e207db656d385c9)) by @bchatard

- *(hooks)* Unit tests for hook template rendering - ([bc9a846](https://github.com/adaouat/bifrost/commit/bc9a846eb201898da717c240996310bd1db7ba02)) by @bchatard

- *(strategy/atomic)* Integration tests for shared linking algorithm - ([ca29e49](https://github.com/adaouat/bifrost/commit/ca29e491ab5040c882f13bc9c9a6ef4b7085f146)) by @bchatard

## [0.7.0](https://github.com/adaouat/bifrost/compare/v0.6.0..v0.7.0) - 2026-05-30

### 🚀 Features

- *(deploy,release)* Dry-run mode prints planned actions without executing - ([31f2b13](https://github.com/adaouat/bifrost/commit/31f2b13090feadef53186a904e13ea5756e3619f)) by @bchatard

- *(release)* Interactive release selector when --release is omitted - ([51499b5](https://github.com/adaouat/bifrost/commit/51499b5e16a6ad899e1977b4df4a11603e1533b0)) by @bchatard

- *(tui)* JSON event stream for deploy and release list - ([d489806](https://github.com/adaouat/bifrost/commit/d489806ccf0f808742b050facf71b7518f3100cc)) by @bchatard

- *(tui)* Add plain output mode — no colors, no spinners - ([c1a89d3](https://github.com/adaouat/bifrost/commit/c1a89d3dbd59920bd8ecec0e89f37d8b84b7ba62)) by @bchatard

- *(tui)* NO_COLOR support and plain-mode color suppression - ([111c7ad](https://github.com/adaouat/bifrost/commit/111c7ade8f282c36bcb077d32a6f0c01ed7b7c7b)) by @bchatard


### 📚 Documentation

- *(roadmap)* Add human mode step output task to M6 - ([253668c](https://github.com/adaouat/bifrost/commit/253668c34fa9a72f9b02112053db75519af32ff3)) by @bchatard

- *(roadmap)* Replace M6 stub with M8 spec-06 completion tasks - ([541d096](https://github.com/adaouat/bifrost/commit/541d0964b128281778fc29e603f29bd10be84bb9)) by @bchatard


### 🧪 Testing

- *(tui)* Verify non-TTY detection for interactive hooks and prompts - ([6f86806](https://github.com/adaouat/bifrost/commit/6f868060459c4dfb54ab9efe756b247b7475ad79)) by @bchatard

## [0.6.0](https://github.com/adaouat/bifrost/compare/v0.5.0..v0.6.0) - 2026-05-30

### 🚀 Features

- *(cmdutil)* Centralise config-path resolution with BIFROST_FILE env var - ([bcb16c1](https://github.com/adaouat/bifrost/commit/bcb16c14c7a9f372f836bc233d865caf2654771f)) by @bchatard

- *(config)* Promote config to subcommand group — add config show - ([e52001a](https://github.com/adaouat/bifrost/commit/e52001a2defbee45ba447e233a7f73fbd59536b5)) by @bchatard

- *(config)* Add config check subcommand - ([6994b39](https://github.com/adaouat/bifrost/commit/6994b39f73a4ccd2e9cb63be49a6abde2ee3cc5c)) by @bchatard

- *(config)* Add config init subcommand - ([45c59da](https://github.com/adaouat/bifrost/commit/45c59da5596179bee67a7b26534c7b91be038f24)) by @bchatard


### 🐛 Bug Fixes

- *(config)* Config check exits 0 silently on valid config - ([d03bfc3](https://github.com/adaouat/bifrost/commit/d03bfc35976324e6880a2d9dd6046d4d9828b496)) by @bchatard

- *(config)* Document that init test globals are not parallel-safe - ([913d891](https://github.com/adaouat/bifrost/commit/913d8915998d73d8364d63b47c540f67d581c0e5)) by @bchatard

- *(config)* Require --env and --app together in config show - ([0bf33d3](https://github.com/adaouat/bifrost/commit/0bf33d31a534936f82af98bbdba66eb8709fbba9)) by @bchatard

- *(config)* Pre-format scaffold to match yamlfmt output - ([76e9504](https://github.com/adaouat/bifrost/commit/76e950441229b9ea27f3803d814c523f171cab9f)) by @bchatard


### 🚜 Refactor

- *(config)* Rename statFile to pathStat for consistency - ([438dbc1](https://github.com/adaouat/bifrost/commit/438dbc1f5ea5501d215a8e3f13a66aec6174f72c)) by @bchatard


### 📚 Documentation

- Update CLAUDE.md and workflow rules after M5 - ([29db012](https://github.com/adaouat/bifrost/commit/29db0127f03df48d1aa2839f4e6f43c70ae2f158)) by @bchatard


### 🧪 Testing

- *(config)* Use assert.Contains over assert.True(strings.Contains) - ([9769d60](https://github.com/adaouat/bifrost/commit/9769d600c74a96d4564f0d2b2b9a50841527c9bf)) by @bchatard


### ⚙️ Miscellaneous Tasks

- Add gopls and Claude Code plugins for LSP and dev tooling - ([4e97c89](https://github.com/adaouat/bifrost/commit/4e97c8993d29729a5efff3f42e4565bc5af47e42)) by @bchatard

- Fix workflow errors and pin actions to SHA - ([cdc1bf4](https://github.com/adaouat/bifrost/commit/cdc1bf48426ed422156fadcda1cf640bee09644d)) by @bchatard

## [0.5.0](https://github.com/adaouat/bifrost/compare/v0.4.0..v0.5.0) - 2026-05-30

### 🚀 Features

- *(cmd)* Implement init command — create releases_root and shared_root - ([211201e](https://github.com/adaouat/bifrost/commit/211201ef4721d9b55a339da2d60e054dc97d61d4)) by @bchatard

- *(cmd/release)* Implement release rollback — activate the preceding release - ([e6affcc](https://github.com/adaouat/bifrost/commit/e6affcc7d9ca8514af62c01dc3f859d40296e9cd)) by @bchatard


### 🚜 Refactor

- *(cmd)* Move init under release — bifrost release init - ([7692ba8](https://github.com/adaouat/bifrost/commit/7692ba896e8beb7f51d784764d0df97483d98b01)) by @bchatard


### 📚 Documentation

- *(roadmap)* Mark --release-name as done — already implemented on deploy cmd - ([54c213a](https://github.com/adaouat/bifrost/commit/54c213af55a0e4a425fcf85f7b84550decbd28c4)) by @bchatard

- *(roadmap)* Add M5 config restructure, renumber M5→M6 and M6→M7 - ([e9937a6](https://github.com/adaouat/bifrost/commit/e9937a6481c4dc00fa8620c45090ffaa2cc5c4d0)) by @bchatard


### ⚙️ Miscellaneous Tasks

- Init Héraut config - ([6056d21](https://github.com/adaouat/bifrost/commit/6056d211a598a1e96c1958de7878ea9ce01e779a)) by @bchatard

- Exclude CHANGELOG.md from typos — commit SHAs trigger false positives - ([7dcc188](https://github.com/adaouat/bifrost/commit/7dcc1884a971cf23678668e0b8001c34feac6915)) by @bchatard

## [0.4.0](https://github.com/adaouat/bifrost/compare/v0.3.0..v0.4.0) - 2026-05-24

### 🚀 Features

- *(cmd/deploy)* Wire hooks into deploy flow at all four lifecycle points - ([67949c4](https://github.com/adaouat/bifrost/commit/67949c4a405002eff67b6b37d68e9e1c015f4af9)) by @bchatard

- *(cmd/release)* Implement release activate — link shared, run hooks, update current - ([4d9550e](https://github.com/adaouat/bifrost/commit/4d9550e56a9931a7ae5205f114ab1cc931c75882)) by @bchatard

- *(cmd/release)* Implement release list — newest-first with active marker - ([fd3a4fa](https://github.com/adaouat/bifrost/commit/fd3a4fa6c0b03ba6ec0dcb29330d48926cd1cf06)) by @bchatard

- *(hooks)* Implement hook runner — sh -c, sudo, template rendering - ([6b4b05c](https://github.com/adaouat/bifrost/commit/6b4b05ca2d9ea9a8ee5370129814ba897a8d121f)) by @bchatard

- *(hooks)* Add cmd_dir support — per-hook working directory override - ([9392b71](https://github.com/adaouat/bifrost/commit/9392b715bc9ae4a42d9a2067598b09e6d11f6355)) by @bchatard

- *(hooks)* Add allow_fail support — warn and continue on non-zero exit - ([1b07a15](https://github.com/adaouat/bifrost/commit/1b07a15b790d0fa5fa17d9d587258ee76b638461)) by @bchatard

- *(hooks)* Add interactive support — confirmFn skips or runs hook on user choice - ([58fbc1c](https://github.com/adaouat/bifrost/commit/58fbc1c4594656ce54ca709969fedf7e53bdb8a7)) by @bchatard

## [0.3.0](https://github.com/adaouat/bifrost/compare/v0.2.0..v0.3.0) - 2026-05-24

### 🚀 Features

- *(cmd/deploy)* Implement deploy command — full flow, no hooks - ([58e32ed](https://github.com/adaouat/bifrost/commit/58e32edc5e9a922a870d1697b6d06efa31180de2)) by @bchatard

- *(cmd/deploy)* Wire progress bar to extraction byte stream - ([c3d0d88](https://github.com/adaouat/bifrost/commit/c3d0d88b3bdb5f5f87951b6b4b6960cc0e53aa72)) by @bchatard

- *(strategy/atomic)* Implement archive extraction - ([f6b72e0](https://github.com/adaouat/bifrost/commit/f6b72e01cb91b6e19cd22c808970c2399498743a)) by @bchatard

- *(strategy/atomic)* Implement shared resource linking - ([2b239f0](https://github.com/adaouat/bifrost/commit/2b239f0ed1f8cd2793185f4ef8db7612b4480969)) by @bchatard

- *(strategy/atomic)* Implement release directory management - ([d5746d5](https://github.com/adaouat/bifrost/commit/d5746d594e62fa242de3a4ae9e8d5b541fb2f798)) by @bchatard

- *(tui)* Add spinner and progress bar wrappers - ([88d825b](https://github.com/adaouat/bifrost/commit/88d825b914a721ebbb066c85ea730be2f8ce73c5)) by @bchatard

## [0.2.0](https://github.com/adaouat/bifrost/compare/v0.0.0..v0.2.0) - 2026-05-23

### 🚀 Features

- *(cmd)* Add fang entry point and root command stub - ([5a2ea3a](https://github.com/adaouat/bifrost/commit/5a2ea3aa0bbf00d521929bd7a946f0718f160f51)) by @bchatard

- *(cmd)* Add global flags --config, --output, --dry-run, --verbose - ([19993ad](https://github.com/adaouat/bifrost/commit/19993adcc35fe1c3764265e73156960d307af9e9)) by @bchatard

- *(cmd)* Add stub commands deploy, config, release list/activate/rollback - ([1456eda](https://github.com/adaouat/bifrost/commit/1456eda969970bb49f9b8dcb983e2a8082bc1a16)) by @bchatard

- *(cmd)* Auto-discover config file (.config/bifrost.yml then .bifrost.yml) - ([37a86f7](https://github.com/adaouat/bifrost/commit/37a86f73e49a22e35b545c688293bbb36668e667)) by @bchatard

- *(cmd/config)* Implement config command — print full config as JSON - ([a066606](https://github.com/adaouat/bifrost/commit/a066606b2ee39eb36bd009624553e174bfb6e8db)) by @bchatard

- *(cmd/config)* Add --env/--app merge, validation, and exit code 2 - ([49cadfd](https://github.com/adaouat/bifrost/commit/49cadfdd8f11c9010f33af4d59ca7454ca81a01d)) by @bchatard

- *(config)* Add schema.go with Go structs for full YAML config - ([f542428](https://github.com/adaouat/bifrost/commit/f54242806c969c0ec5c5639ed505afc80cc96cb0)) by @bchatard

- *(config)* Add strict YAML loader with defaults - ([1c2fd2b](https://github.com/adaouat/bifrost/commit/1c2fd2bbdd913031b12659cd3e6f8f7733b94e70)) by @bchatard

- *(config)* Add three-level merge (global < env < app) - ([cd69327](https://github.com/adaouat/bifrost/commit/cd69327938cd202c3378faab342240d0ab4da73a)) by @bchatard

- *(tui)* Add lipgloss styles and ANSI shadow header for help/version - ([51aafa4](https://github.com/adaouat/bifrost/commit/51aafa42e34fc47418c480dcb3d2999fbca01c25)) by @bchatard


### 📚 Documentation

- *(hooks)* Clarify distinction between pre_link and pre_enable_release - ([5b338ec](https://github.com/adaouat/bifrost/commit/5b338ec347d8cf495f2ed8b1a6995039bb712fed)) by @bchatard

- *(roadmap)* Add working process - ([5ddc8c0](https://github.com/adaouat/bifrost/commit/5ddc8c0ea37abfc216786c1cad7b63e3a0a80ca6)) by @bchatard

- *(roadmap)* Mark M0 tooling task 1 done (mise Go/golangci-lint/git-cliff) - ([8016144](https://github.com/adaouat/bifrost/commit/80161446859b6d74d932a59f604a1ca49e7324d5)) by @bchatard

- Add CLAUDE.md, specs, ADRs, and development roadmap - ([e07ef64](https://github.com/adaouat/bifrost/commit/e07ef64ca1a030e361a1f8e2384892e008778d2a)) by @bchatard

- Name the tool Bifrost (ADR-0005) - ([6329883](https://github.com/adaouat/bifrost/commit/6329883dac6521c75e465a2b230affc445e2e5d5)) by @bchatard

- Rename deployer → bifrost across all docs - ([ab1f6cd](https://github.com/adaouat/bifrost/commit/ab1f6cd240333485a00db393a7ec4540eb72faad)) by @bchatard

- Define coding strategy and expand M0 with CI/CD - ([b6a4736](https://github.com/adaouat/bifrost/commit/b6a47367ee5c642df7e5a18949ff6d6fc6c9a06d)) by @bchatard

- Introduce strategy-based deployment architecture (ADR-0006) - ([42c0896](https://github.com/adaouat/bifrost/commit/42c0896d33f789858e04f0bb9a29a532c59e8e51)) by @bchatard

- Add version roadmap (v0–v5) capturing design decisions - ([c0987d5](https://github.com/adaouat/bifrost/commit/c0987d55371b6e47cf49bc6a47e5403bd76e9560)) by @bchatard

- Add testing strategy (ADR-0007) - ([104d806](https://github.com/adaouat/bifrost/commit/104d806bb9a32494375b66d9287ab7c3bad21b1f)) by @bchatard

- Rename artifact strategy to atomic, replace Capistrano-style - ([d1bfe4f](https://github.com/adaouat/bifrost/commit/d1bfe4f21530f53d136c694ffb700240e22e7f39)) by @bchatard

- Final sweep before implementation sessions - ([43e5f5a](https://github.com/adaouat/bifrost/commit/43e5f5a7255a6e24cf8b80338131fbb0112a8758)) by @bchatard


### ⚙️ Miscellaneous Tasks

- *(claude)* Store plans in project, migrate command structure plan - ([f5a2551](https://github.com/adaouat/bifrost/commit/f5a2551de47ca11364c93ad18fc9b276dacc15a3)) by @bchatard

- *(hk)* Add go_fmt, golangci_lint, gomod_tidy linters - ([8879238](https://github.com/adaouat/bifrost/commit/8879238a52fd686d54e0ba158a5c50868963bb9f)) by @bchatard

- *(mise)* Add build, test, run, lint:go:check, lint:go:fix tasks - ([af99a3f](https://github.com/adaouat/bifrost/commit/af99a3f0ce0ec37a8916979337648b700ec7cc75)) by @bchatard

- *(testdata)* Add YAML config fixtures and release archive - ([b283f48](https://github.com/adaouat/bifrost/commit/b283f486d344f6df85891f4a9c58bb49c3139436)) by @bchatard

- *(testutil)* Add container test helpers and binary builder - ([42e9e7f](https://github.com/adaouat/bifrost/commit/42e9e7f829229c00a6ad146f22ef958967f5fbb6)) by @bchatard

- Init project - ([71743b6](https://github.com/adaouat/bifrost/commit/71743b6213536fca37199ba09edb4944b5a59df9)) by @bchatard

- Add Claude rules and restructure CLAUDE.md - ([1cbab33](https://github.com/adaouat/bifrost/commit/1cbab33c22a0710dc8f205d8a8e572197f5b81ae)) by @bchatard

- Commit Claude settings, gitignore, and roadmap update - ([9daff9b](https://github.com/adaouat/bifrost/commit/9daff9b246479efdfc1baeb9fdee38a8075228e7)) by @bchatard

- Pin tool versions, add cliff.toml, clean up cocogitto - ([c4e9a08](https://github.com/adaouat/bifrost/commit/c4e9a0857b171d60f1600fdb36b0ff2d608b76fb)) by @bchatard

- Go mod init github.com/bchatard/bifrost - ([2388f29](https://github.com/adaouat/bifrost/commit/2388f29a48638ad9dcc57bb7df5560f658332e8e)) by @bchatard

- Rename module to github.com/adaouat/bifrost - ([c814287](https://github.com/adaouat/bifrost/commit/c81428795566c207b463e59886136102821748e2)) by @bchatard

- Add goreleaser config and GitHub Actions workflows - ([9aa2e8d](https://github.com/adaouat/bifrost/commit/9aa2e8dc105b28c8830115220e6618a2601f30d6)) by @bchatard

