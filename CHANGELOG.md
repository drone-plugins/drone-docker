# Changelog

## [v20.14.5](https://github.com/drone-plugins/drone-docker/tree/v20.14.5) (2023-09-13)

[Full Changelog](https://github.com/drone-plugins/drone-docker/compare/v20.14.4...v20.14.5)

**Implemented enhancements:**

- Allow gcr authentication with workload identity [\#383](https://github.com/drone-plugins/drone-docker/pull/383) ([dhpollack](https://github.com/dhpollack))

**Fixed bugs:**

- \[fix\]: \[ci-9254\]: go version upgrade to 1.21 [\#401](https://github.com/drone-plugins/drone-docker/pull/401) ([abhay084](https://github.com/abhay084))
- Revert "Add support for AAD auth for docker-acr" [\#398](https://github.com/drone-plugins/drone-docker/pull/398) ([tphoney](https://github.com/tphoney))

**Closed issues:**

- Remove deprecated support of label-schema in favor of OCI [\#396](https://github.com/drone-plugins/drone-docker/issues/396)

**Merged pull requests:**

- Add support for AAD auth for docker-acr [\#395](https://github.com/drone-plugins/drone-docker/pull/395) ([rutvijmehta-harness](https://github.com/rutvijmehta-harness))

## [v20.14.4](https://github.com/drone-plugins/drone-docker/tree/v20.14.4) (2023-05-16)

[Full Changelog](https://github.com/drone-plugins/drone-docker/compare/v20.14.3...v20.14.4)

**Fixed bugs:**

- fix: Use unique build name for build and tag [\#390](https://github.com/drone-plugins/drone-docker/pull/390) ([rutvijmehta-harness](https://github.com/rutvijmehta-harness))

**Merged pull requests:**

- v20.14.4 prep [\#391](https://github.com/drone-plugins/drone-docker/pull/391) ([rutvijmehta-harness](https://github.com/rutvijmehta-harness))

## [v20.14.3](https://github.com/drone-plugins/drone-docker/tree/v20.14.3) (2023-05-04)

[Full Changelog](https://github.com/drone-plugins/drone-docker/compare/v20.14.2...v20.14.3)

**Merged pull requests:**

- Write artifacts to input artifact file [\#389](https://github.com/drone-plugins/drone-docker/pull/389) ([raghavharness](https://github.com/raghavharness))

## [v20.14.2](https://github.com/drone-plugins/drone-docker/tree/v20.14.2) (2023-04-18)

[Full Changelog](https://github.com/drone-plugins/drone-docker/compare/v20.14.1...v20.14.2)

**Merged pull requests:**

- fix windows drone yml [\#388](https://github.com/drone-plugins/drone-docker/pull/388) ([shubham149](https://github.com/shubham149))

## [v20.14.1](https://github.com/drone-plugins/drone-docker/tree/v20.14.1) (2023-01-30)

[Full Changelog](https://github.com/drone-plugins/drone-docker/compare/v20.14.0...v20.14.1)

**Implemented enhancements:**

- Add option to mount host ssh agent \(--ssh\) [\#382](https://github.com/drone-plugins/drone-docker/pull/382) ([tphoney](https://github.com/tphoney))

**Fixed bugs:**

- windows 1809 docker build pin [\#384](https://github.com/drone-plugins/drone-docker/pull/384) ([tphoney](https://github.com/tphoney))
- \(maint\) move to harness.drone.io [\#381](https://github.com/drone-plugins/drone-docker/pull/381) ([tphoney](https://github.com/tphoney))

## [v20.14.0](https://github.com/drone-plugins/drone-docker/tree/v20.14.0) (2022-11-17)

[Full Changelog](https://github.com/drone-plugins/drone-docker/compare/v20.13.0...v20.14.0)

**Implemented enhancements:**

- Add support for docker --platform flag [\#376](https://github.com/drone-plugins/drone-docker/pull/376) ([tphoney](https://github.com/tphoney))

**Fixed bugs:**

- Use full path to docker when creating card [\#373](https://github.com/drone-plugins/drone-docker/pull/373) ([donny-dont](https://github.com/donny-dont))

**Merged pull requests:**

- \(maint\) prep for v20.14.0 [\#377](https://github.com/drone-plugins/drone-docker/pull/377) ([tphoney](https://github.com/tphoney))

## [v20.13.0](https://github.com/drone-plugins/drone-docker/tree/v20.13.0) (2022-06-08)

[Full Changelog](https://github.com/drone-plugins/drone-docker/compare/v20.12.0...v20.13.0)

**Implemented enhancements:**

- update docker linux amd64/arm64 to 20.10.14 [\#365](https://github.com/drone-plugins/drone-docker/pull/365) ([tphoney](https://github.com/tphoney))

**Merged pull requests:**

- v20.13.0 prep [\#367](https://github.com/drone-plugins/drone-docker/pull/367) ([tphoney](https://github.com/tphoney))

## [v20.12.0](https://github.com/drone-plugins/drone-docker/tree/v20.12.0) (2022-05-16)

[Full Changelog](https://github.com/drone-plugins/drone-docker/compare/v20.11.0...v20.12.0)

**Implemented enhancements:**

- Add support for multiple Buildkit secrets with env vars or files as source [\#359](https://github.com/drone-plugins/drone-docker/pull/359) ([ste93cry](https://github.com/ste93cry))
- \(DRON-237\) cards add link to image repo, minor cleanup [\#358](https://github.com/drone-plugins/drone-docker/pull/358) ([tphoney](https://github.com/tphoney))
- \(DRON-232\) enable build-kit for secrets consumption [\#356](https://github.com/drone-plugins/drone-docker/pull/356) ([tphoney](https://github.com/tphoney))

**Fixed bugs:**

- \(fix\) Update card.json with UX [\#355](https://github.com/drone-plugins/drone-docker/pull/355) ([tphoney](https://github.com/tphoney))

**Merged pull requests:**

- prep for v20.12.0 [\#363](https://github.com/drone-plugins/drone-docker/pull/363) ([tphoney](https://github.com/tphoney))

## [v20.11.0](https://github.com/drone-plugins/drone-docker/tree/v20.11.0) (2022-01-19)

[Full Changelog](https://github.com/drone-plugins/drone-docker/compare/v20.10.9.1...v20.11.0)

**Merged pull requests:**

- new release to fix window semver error [\#354](https://github.com/drone-plugins/drone-docker/pull/354) ([eoinmcafee00](https://github.com/eoinmcafee00))
- \(feat\) publish docker data to create drone card [\#347](https://github.com/drone-plugins/drone-docker/pull/347) ([eoinmcafee00](https://github.com/eoinmcafee00))

## [v20.10.9.1](https://github.com/drone-plugins/drone-docker/tree/v20.10.9.1) (2022-01-13)

[Full Changelog](https://github.com/drone-plugins/drone-docker/compare/v20.10.9...v20.10.9.1)

**Implemented enhancements:**

- Serialize windows 1809 pipelines [\#348](https://github.com/drone-plugins/drone-docker/pull/348) ([shubham149](https://github.com/shubham149))
- Support for windows images for tags [\#346](https://github.com/drone-plugins/drone-docker/pull/346) ([shubham149](https://github.com/shubham149))

**Fixed bugs:**

- Fix ECR & GCR docker publish on windows [\#352](https://github.com/drone-plugins/drone-docker/pull/352) ([shubham149](https://github.com/shubham149))
- Fix windows docker builds [\#351](https://github.com/drone-plugins/drone-docker/pull/351) ([shubham149](https://github.com/shubham149))
- Fix powershell script to publish windows images [\#350](https://github.com/drone-plugins/drone-docker/pull/350) ([shubham149](https://github.com/shubham149))

**Merged pull requests:**

- release prep for 20.10.9.1 [\#353](https://github.com/drone-plugins/drone-docker/pull/353) ([eoinmcafee00](https://github.com/eoinmcafee00))

## [v20.10.9](https://github.com/drone-plugins/drone-docker/tree/v20.10.9) (2021-11-03)

[Full Changelog](https://github.com/drone-plugins/drone-docker/compare/v19.03.9...v20.10.9)

**Merged pull requests:**

- bump to version 20.10.9: [\#342](https://github.com/drone-plugins/drone-docker/pull/342) ([eoinmcafee00](https://github.com/eoinmcafee00))
- Upgrade Docker dind to 20.10.9 for 64bit platforms [\#334](https://github.com/drone-plugins/drone-docker/pull/334) ([gzm0](https://github.com/gzm0))

## [v19.03.9](https://github.com/drone-plugins/drone-docker/tree/v19.03.9) (2021-10-13)

[Full Changelog](https://github.com/drone-plugins/drone-docker/compare/v19.03.8...v19.03.9)

**Implemented enhancements:**

- adding support for externalId [\#333](https://github.com/drone-plugins/drone-docker/pull/333) ([jimsheldon](https://github.com/jimsheldon))
- Add support for automatic opencontainer labels [\#313](https://github.com/drone-plugins/drone-docker/pull/313) ([codrut-fc](https://github.com/codrut-fc))
- add custom seccomp profile [\#312](https://github.com/drone-plugins/drone-docker/pull/312) ([xoxys](https://github.com/xoxys))
- ECR: adding setting to enable image scanning while repo creation [\#300](https://github.com/drone-plugins/drone-docker/pull/300) ([rvoitenko](https://github.com/rvoitenko))

**Fixed bugs:**

- Revert "Update seccomp to 20.10 docker" [\#325](https://github.com/drone-plugins/drone-docker/pull/325) ([bradrydzewski](https://github.com/bradrydzewski))

**Closed issues:**

- Enable auth against multiple registries [\#324](https://github.com/drone-plugins/drone-docker/issues/324)
- Parameter add\_host not work [\#318](https://github.com/drone-plugins/drone-docker/issues/318)
- support customized Dockerfile name ? [\#315](https://github.com/drone-plugins/drone-docker/issues/315)
- Tag wrongly gets parsed as octal [\#311](https://github.com/drone-plugins/drone-docker/issues/311)
- Support TLS 1.3 [\#310](https://github.com/drone-plugins/drone-docker/issues/310)
- Can pugin-docker access workspace content directly?  [\#307](https://github.com/drone-plugins/drone-docker/issues/307)

**Merged pull requests:**

- \(maint\) bump git to 1.13 for build and test [\#338](https://github.com/drone-plugins/drone-docker/pull/338) ([tphoney](https://github.com/tphoney))
- \(maint\) v19.03.9 release prep [\#337](https://github.com/drone-plugins/drone-docker/pull/337) ([tphoney](https://github.com/tphoney))
- \(maint\) CI, remove the dry run steps, due to rate limiting [\#323](https://github.com/drone-plugins/drone-docker/pull/323) ([tphoney](https://github.com/tphoney))
- Update seccomp to 20.10 docker [\#322](https://github.com/drone-plugins/drone-docker/pull/322) ([techknowlogick](https://github.com/techknowlogick))



\* *This Changelog was automatically generated by [github_changelog_generator](https://github.com/github-changelog-generator/github-changelog-generator)*
