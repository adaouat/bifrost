# Changelog

## [0.10.0](//compare/v0.9.0..v0.10.0) - 2026-06-02

### 🚀 Features

- *(cmd)* Add --agent-binary flag to deploy and SSH-capable release commands - ([98fc9c9](//commit/98fc9c9109f75d4fc9acfa9aac7c0838c696e47a))

- *(transport)* Add SSH client wrapper with auth chain and strict host keys - ([4dfeb0b](//commit/4dfeb0b8a5c0b3cb53d625095baebda2016b0c1a))

- *(transport)* Add SFTP wrapper for upload, mkdir, chmod - ([6f98610](//commit/6f986103880cf71437647b4697513fd29a44990c))

- *(transport)* Add remote staging directory lifecycle - ([6e245aa](//commit/6e245aabfbc2c7c4f82114eec9c7445253b2a238))

- *(transport)* Add agent binary resolution with arch detection and cache - ([75d94f2](//commit/75d94f22beb7e9c6e39e4d3a9201ae108bece73c))

## [0.9.0](//compare/v0.8.1..v0.9.0) - 2026-06-02

### 🚀 Features

- *(config)* Add server config schema, validation, and merge resolution (M12) - ([81d729d](//commit/81d729dc1bb39df15fa6728106bffb17f07b48d9)) by @bchatard


### 📚 Documentation

- *(v1)* Add JSON Schema task to M17 - ([56297c7](//commit/56297c75b8341a1856d62a58965981644b2550cf)) by @bchatard

## [0.8.1](//compare/v0.8.0..v0.8.1) - 2026-06-01

### 🚜 Refactor

- *(strategy)* Introduce Deployer interface and move atomic deploy logic - ([08f918b](//commit/08f918b3389da4e8e917e0be58884ed05b2e24d1)) by @bchatard


### 📚 Documentation

- *(v1)* Add roadmap, specs, and ADRs for SSH orchestration + agent model - ([2e8f923](//commit/2e8f9233e37ab0325c3a0da65458213dda4201d4)) by @bchatard

## [0.8.0](//compare/v0.7.1..v0.8.0) - 2026-05-31

### 🚀 Features

- *(cmd/deploy)* Emit JSON error event on step failure - ([9243eb7](//commit/9243eb72247b0361040f16945117ffaab6be1185)) by @bchatard

- *(cmd/deploy)* Start/done JSON events for link, current_symlink, purge - ([934dc89](//commit/934dc89a9c4f44f29c2da7238f8dbcd8747b7c1a)) by @bchatard

- *(cmd/deploy)* Dry-run Would purge line with actual release names - ([bcc62c6](//commit/bcc62c6fb5c043ee7c52805a24432cbab58ef1bc)) by @bchatard

- *(cmd/deploy)* Show (sudo) suffix for sudo hooks in dry-run output - ([1d69311](//commit/1d693118ad5740e5192564099502568ad025ba51)) by @bchatard

- *(cmd/deploy)* Spinner while purging old releases - ([d212288](//commit/d212288285c518e2422aa5a2dd9727172342486e)) by @bchatard

- *(release/list)* Header line and ← current suffix in human output - ([94f7a06](//commit/94f7a0606c1b23b3adb00a0ed640f0e74462751c)) by @bchatard

- *(tui)* Deploy header panel with bordered box - ([d4b5404](//commit/d4b5404e943ea5a4817a8b0fa2f96410fa9057bd)) by @bchatard

- *(tui)* Per-step checkmark lines for deploy human output - ([1ab7ca6](//commit/1ab7ca6aa2506cd3020a47ffeaef451a904335be)) by @bchatard

- *(tui)* Final summary line for deploy human output - ([5429429](//commit/54294296ae148fd2fbfcd2898a2512695f70e3d9)) by @bchatard

- *(tui)* Step detail sub-lines and enriched JSON events - ([71cf30a](//commit/71cf30ad29181bc136c5d6faa4f64847f79ff409)) by @bchatard

- *(tui)* Add title prefix to extraction progress bar - ([e381a84](//commit/e381a843d18219224d5892de8abe2bd76fc8ba54)) by @bchatard


### 🐛 Bug Fixes

- *(strategy/atomic)* Track compressed bytes for extraction progress - ([58de9e2](//commit/58de9e2f17b0631520c337f9585efe461a8cd93b)) by @bchatard

- *(strategy/atomic)* Protect active release from purge - ([3386860](//commit/33868600c4d733930e233ad257f850a385e84932)) by @bchatard

- Address important review findings - ([4143c38](//commit/4143c3878a76c5e762c3924d5b62318ab9440a52)) by @bchatard


### 🚜 Refactor

- Address suggestion review findings - ([8873db2](//commit/8873db2c665cf84ac64821ec11ead5248a41c1f5)) by @bchatard


### 📚 Documentation

- *(claude)* Sync CLAUDE.md project layout with current codebase - ([95b9018](//commit/95b9018f6711737be17ecfb58515ccd053a34a66)) by @bchatard

- *(specs)* Sync 06-tui-ux and 03-commands with current implementation - ([5fa71ef](//commit/5fa71efefcf614bf173d6c10e04b906474289018)) by @bchatard

## [0.7.1](//compare/v0.7.0..v0.7.1) - 2026-05-31

### 🧪 Testing

- *(cmd/deploy)* Integration test for release purge - ([7ccb14a](//commit/7ccb14a410358dd0123fa6163df093f822a8bfc7)) by @bchatard

- *(cmd/release)* Integration test for activating non-existent release - ([b80ace3](//commit/b80ace33d6186547d26397f607583a481620af96)) by @bchatard

- *(config)* Unit tests for merge logic - ([35b07a3](//commit/35b07a301c2430bdc48ad8153e207db656d385c9)) by @bchatard

- *(hooks)* Unit tests for hook template rendering - ([bc9a846](//commit/bc9a846eb201898da717c240996310bd1db7ba02)) by @bchatard

- *(strategy/atomic)* Integration tests for shared linking algorithm - ([ca29e49](//commit/ca29e491ab5040c882f13bc9c9a6ef4b7085f146)) by @bchatard

## [0.7.0](//compare/v0.6.0..v0.7.0) - 2026-05-30

### 🚀 Features

- *(deploy,release)* Dry-run mode prints planned actions without executing - ([31f2b13](//commit/31f2b13090feadef53186a904e13ea5756e3619f)) by @bchatard

- *(release)* Interactive release selector when --release is omitted - ([51499b5](//commit/51499b5e16a6ad899e1977b4df4a11603e1533b0)) by @bchatard

- *(tui)* JSON event stream for deploy and release list - ([d489806](//commit/d489806ccf0f808742b050facf71b7518f3100cc)) by @bchatard

- *(tui)* Add plain output mode — no colors, no spinners - ([c1a89d3](//commit/c1a89d3dbd59920bd8ecec0e89f37d8b84b7ba62)) by @bchatard

- *(tui)* NO_COLOR support and plain-mode color suppression - ([111c7ad](//commit/111c7ade8f282c36bcb077d32a6f0c01ed7b7c7b)) by @bchatard


### 📚 Documentation

- *(roadmap)* Add human mode step output task to M6 - ([253668c](//commit/253668c34fa9a72f9b02112053db75519af32ff3)) by @bchatard

- *(roadmap)* Replace M6 stub with M8 spec-06 completion tasks - ([541d096](//commit/541d0964b128281778fc29e603f29bd10be84bb9)) by @bchatard


### 🧪 Testing

- *(tui)* Verify non-TTY detection for interactive hooks and prompts - ([6f86806](//commit/6f868060459c4dfb54ab9efe756b247b7475ad79)) by @bchatard

## [0.6.0](//compare/v0.5.0..v0.6.0) - 2026-05-30

### 🚀 Features

- *(cmdutil)* Centralise config-path resolution with BIFROST_FILE env var - ([bcb16c1](//commit/bcb16c14c7a9f372f836bc233d865caf2654771f)) by @bchatard

- *(config)* Promote config to subcommand group — add config show - ([e52001a](//commit/e52001a2defbee45ba447e233a7f73fbd59536b5)) by @bchatard

- *(config)* Add config check subcommand - ([6994b39](//commit/6994b39f73a4ccd2e9cb63be49a6abde2ee3cc5c)) by @bchatard

- *(config)* Add config init subcommand - ([45c59da](//commit/45c59da5596179bee67a7b26534c7b91be038f24)) by @bchatard


### 🐛 Bug Fixes

- *(config)* Config check exits 0 silently on valid config - ([d03bfc3](//commit/d03bfc35976324e6880a2d9dd6046d4d9828b496)) by @bchatard

- *(config)* Document that init test globals are not parallel-safe - ([913d891](//commit/913d8915998d73d8364d63b47c540f67d581c0e5)) by @bchatard

- *(config)* Require --env and --app together in config show - ([0bf33d3](//commit/0bf33d31a534936f82af98bbdba66eb8709fbba9)) by @bchatard

- *(config)* Pre-format scaffold to match yamlfmt output - ([76e9504](//commit/76e950441229b9ea27f3803d814c523f171cab9f)) by @bchatard


### 🚜 Refactor

- *(config)* Rename statFile to pathStat for consistency - ([438dbc1](//commit/438dbc1f5ea5501d215a8e3f13a66aec6174f72c)) by @bchatard


### 📚 Documentation

- Update CLAUDE.md and workflow rules after M5 - ([29db012](//commit/29db0127f03df48d1aa2839f4e6f43c70ae2f158)) by @bchatard


### 🧪 Testing

- *(config)* Use assert.Contains over assert.True(strings.Contains) - ([9769d60](//commit/9769d600c74a96d4564f0d2b2b9a50841527c9bf)) by @bchatard


### ⚙️ Miscellaneous Tasks

- Add gopls and Claude Code plugins for LSP and dev tooling - ([4e97c89](//commit/4e97c8993d29729a5efff3f42e4565bc5af47e42)) by @bchatard

- Fix workflow errors and pin actions to SHA - ([cdc1bf4](//commit/cdc1bf48426ed422156fadcda1cf640bee09644d)) by @bchatard

## [0.5.0](//compare/v0.4.0..v0.5.0) - 2026-05-30

### 🚀 Features

- *(cmd)* Implement init command — create releases_root and shared_root - ([211201e](//commit/211201ef4721d9b55a339da2d60e054dc97d61d4)) by @bchatard

- *(cmd/release)* Implement release rollback — activate the preceding release - ([e6affcc](//commit/e6affcc7d9ca8514af62c01dc3f859d40296e9cd)) by @bchatard


### 🚜 Refactor

- *(cmd)* Move init under release — bifrost release init - ([7692ba8](//commit/7692ba896e8beb7f51d784764d0df97483d98b01)) by @bchatard


### 📚 Documentation

- *(roadmap)* Mark --release-name as done — already implemented on deploy cmd - ([54c213a](//commit/54c213af55a0e4a425fcf85f7b84550decbd28c4)) by @bchatard

- *(roadmap)* Add M5 config restructure, renumber M5→M6 and M6→M7 - ([e9937a6](//commit/e9937a6481c4dc00fa8620c45090ffaa2cc5c4d0)) by @bchatard


### ⚙️ Miscellaneous Tasks

- Init Héraut config - ([6056d21](//commit/6056d211a598a1e96c1958de7878ea9ce01e779a)) by @bchatard

- Exclude CHANGELOG.md from typos — commit SHAs trigger false positives - ([7dcc188](//commit/7dcc1884a971cf23678668e0b8001c34feac6915)) by @bchatard

## [0.4.0](//compare/v0.3.0..v0.4.0) - 2026-05-24

### 🚀 Features

- *(cmd/deploy)* Wire hooks into deploy flow at all four lifecycle points - ([67949c4](//commit/67949c4a405002eff67b6b37d68e9e1c015f4af9)) by @bchatard

- *(cmd/release)* Implement release activate — link shared, run hooks, update current - ([4d9550e](//commit/4d9550e56a9931a7ae5205f114ab1cc931c75882)) by @bchatard

- *(cmd/release)* Implement release list — newest-first with active marker - ([fd3a4fa](//commit/fd3a4fa6c0b03ba6ec0dcb29330d48926cd1cf06)) by @bchatard

- *(hooks)* Implement hook runner — sh -c, sudo, template rendering - ([6b4b05c](//commit/6b4b05ca2d9ea9a8ee5370129814ba897a8d121f)) by @bchatard

- *(hooks)* Add cmd_dir support — per-hook working directory override - ([9392b71](//commit/9392b715bc9ae4a42d9a2067598b09e6d11f6355)) by @bchatard

- *(hooks)* Add allow_fail support — warn and continue on non-zero exit - ([1b07a15](//commit/1b07a15b790d0fa5fa17d9d587258ee76b638461)) by @bchatard

- *(hooks)* Add interactive support — confirmFn skips or runs hook on user choice - ([58fbc1c](//commit/58fbc1c4594656ce54ca709969fedf7e53bdb8a7)) by @bchatard

## [0.3.0](//compare/v0.2.0..v0.3.0) - 2026-05-24

### 🚀 Features

- *(cmd/deploy)* Implement deploy command — full flow, no hooks - ([58e32ed](//commit/58e32edc5e9a922a870d1697b6d06efa31180de2)) by @bchatard

- *(cmd/deploy)* Wire progress bar to extraction byte stream - ([c3d0d88](//commit/c3d0d88b3bdb5f5f87951b6b4b6960cc0e53aa72)) by @bchatard

- *(strategy/atomic)* Implement archive extraction - ([f6b72e0](//commit/f6b72e01cb91b6e19cd22c808970c2399498743a)) by @bchatard

- *(strategy/atomic)* Implement shared resource linking - ([2b239f0](//commit/2b239f0ed1f8cd2793185f4ef8db7612b4480969)) by @bchatard

- *(strategy/atomic)* Implement release directory management - ([d5746d5](//commit/d5746d594e62fa242de3a4ae9e8d5b541fb2f798)) by @bchatard

- *(tui)* Add spinner and progress bar wrappers - ([88d825b](//commit/88d825b914a721ebbb066c85ea730be2f8ce73c5)) by @bchatard

## [0.2.0](//compare/v0.0.0..v0.2.0) - 2026-05-23

### 🚀 Features

- *(cmd)* Add fang entry point and root command stub - ([5a2ea3a](//commit/5a2ea3aa0bbf00d521929bd7a946f0718f160f51)) by @bchatard

- *(cmd)* Add global flags --config, --output, --dry-run, --verbose - ([19993ad](//commit/19993adcc35fe1c3764265e73156960d307af9e9)) by @bchatard

- *(cmd)* Add stub commands deploy, config, release list/activate/rollback - ([1456eda](//commit/1456eda969970bb49f9b8dcb983e2a8082bc1a16)) by @bchatard

- *(cmd)* Auto-discover config file (.config/bifrost.yml then .bifrost.yml) - ([37a86f7](//commit/37a86f73e49a22e35b545c688293bbb36668e667)) by @bchatard

- *(cmd/config)* Implement config command — print full config as JSON - ([a066606](//commit/a066606b2ee39eb36bd009624553e174bfb6e8db)) by @bchatard

- *(cmd/config)* Add --env/--app merge, validation, and exit code 2 - ([49cadfd](//commit/49cadfdd8f11c9010f33af4d59ca7454ca81a01d)) by @bchatard

- *(config)* Add schema.go with Go structs for full YAML config - ([f542428](//commit/f54242806c969c0ec5c5639ed505afc80cc96cb0)) by @bchatard

- *(config)* Add strict YAML loader with defaults - ([1c2fd2b](//commit/1c2fd2bbdd913031b12659cd3e6f8f7733b94e70)) by @bchatard

- *(config)* Add three-level merge (global < env < app) - ([cd69327](//commit/cd69327938cd202c3378faab342240d0ab4da73a)) by @bchatard

- *(tui)* Add lipgloss styles and ANSI shadow header for help/version - ([51aafa4](//commit/51aafa42e34fc47418c480dcb3d2999fbca01c25)) by @bchatard


### 📚 Documentation

- *(hooks)* Clarify distinction between pre_link and pre_enable_release - ([5b338ec](//commit/5b338ec347d8cf495f2ed8b1a6995039bb712fed)) by @bchatard

- *(roadmap)* Add working process - ([5ddc8c0](//commit/5ddc8c0ea37abfc216786c1cad7b63e3a0a80ca6)) by @bchatard

- *(roadmap)* Mark M0 tooling task 1 done (mise Go/golangci-lint/git-cliff) - ([8016144](//commit/80161446859b6d74d932a59f604a1ca49e7324d5)) by @bchatard

- Add CLAUDE.md, specs, ADRs, and development roadmap - ([e07ef64](//commit/e07ef64ca1a030e361a1f8e2384892e008778d2a)) by @bchatard

- Name the tool Bifrost (ADR-0005) - ([6329883](//commit/6329883dac6521c75e465a2b230affc445e2e5d5)) by @bchatard

- Rename deployer → bifrost across all docs - ([ab1f6cd](//commit/ab1f6cd240333485a00db393a7ec4540eb72faad)) by @bchatard

- Define coding strategy and expand M0 with CI/CD - ([b6a4736](//commit/b6a47367ee5c642df7e5a18949ff6d6fc6c9a06d)) by @bchatard

- Introduce strategy-based deployment architecture (ADR-0006) - ([42c0896](//commit/42c0896d33f789858e04f0bb9a29a532c59e8e51)) by @bchatard

- Add version roadmap (v0–v5) capturing design decisions - ([c0987d5](//commit/c0987d55371b6e47cf49bc6a47e5403bd76e9560)) by @bchatard

- Add testing strategy (ADR-0007) - ([104d806](//commit/104d806bb9a32494375b66d9287ab7c3bad21b1f)) by @bchatard

- Rename artifact strategy to atomic, replace Capistrano-style - ([d1bfe4f](//commit/d1bfe4f21530f53d136c694ffb700240e22e7f39)) by @bchatard

- Final sweep before implementation sessions - ([43e5f5a](//commit/43e5f5a7255a6e24cf8b80338131fbb0112a8758)) by @bchatard


### ⚙️ Miscellaneous Tasks

- *(claude)* Store plans in project, migrate command structure plan - ([f5a2551](//commit/f5a2551de47ca11364c93ad18fc9b276dacc15a3)) by @bchatard

- *(hk)* Add go_fmt, golangci_lint, gomod_tidy linters - ([8879238](//commit/8879238a52fd686d54e0ba158a5c50868963bb9f)) by @bchatard

- *(mise)* Add build, test, run, lint:go:check, lint:go:fix tasks - ([af99a3f](//commit/af99a3f0ce0ec37a8916979337648b700ec7cc75)) by @bchatard

- *(testdata)* Add YAML config fixtures and release archive - ([b283f48](//commit/b283f486d344f6df85891f4a9c58bb49c3139436)) by @bchatard

- *(testutil)* Add container test helpers and binary builder - ([42e9e7f](//commit/42e9e7f829229c00a6ad146f22ef958967f5fbb6)) by @bchatard

- Init project - ([71743b6](//commit/71743b6213536fca37199ba09edb4944b5a59df9)) by @bchatard

- Add Claude rules and restructure CLAUDE.md - ([1cbab33](//commit/1cbab33c22a0710dc8f205d8a8e572197f5b81ae)) by @bchatard

- Commit Claude settings, gitignore, and roadmap update - ([9daff9b](//commit/9daff9b246479efdfc1baeb9fdee38a8075228e7)) by @bchatard

- Pin tool versions, add cliff.toml, clean up cocogitto - ([c4e9a08](//commit/c4e9a0857b171d60f1600fdb36b0ff2d608b76fb)) by @bchatard

- Go mod init github.com/bchatard/bifrost - ([2388f29](//commit/2388f29a48638ad9dcc57bb7df5560f658332e8e)) by @bchatard

- Rename module to github.com/adaouat/bifrost - ([c814287](//commit/c81428795566c207b463e59886136102821748e2)) by @bchatard

- Add goreleaser config and GitHub Actions workflows - ([9aa2e8d](//commit/9aa2e8dc105b28c8830115220e6618a2601f30d6)) by @bchatard

