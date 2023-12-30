# Changelogs

Also seen in [GitHub Releases](https://github.com/juicity/juicity/releases)

## Query history releases

```bash
curl --silent "https://api.github.com/repos/juicity/juicity/releases" | jq -r '.[] | {tag_name,created_at,release}'
```

## Releases

<!-- BEGIN NEW TOC ENTRY -->

- [v0.4.0rc1 (Pre-release)](#v040rc1-pre-release)
- [v0.3.0 (Latest)](#v030-latest)
- [v0.2.1](#v021)
- [v0.2.0](#v020)
- [v0.1.3](#v013)
- [v0.1.2](#v012)
- [v0.1.1](#v011)
- [v0.1.0](#v010)
<!-- BEGIN NEW CHANGELOGS -->

### v0.4.0rc1 (Pre-release)

> Release date: 2023/12/30

### Features

- feat(cmd/version): add copy right and license info in [#134](https://github.com/juicity/juicity/pull/134) by (@yqlbu)
- feat(cmd): enhance version print in [#123](https://github.com/juicity/juicity/pull/123) by (@yqlbu)

### Others

- ci(lint): upgrade linter ver. in [#142](https://github.com/juicity/juicity/pull/142) by (@sumire88)
- chore: upgrade softwind and quic-go to v0.40.1 in [#141](https://github.com/juicity/juicity/pull/141) by (@mzz2017)
- ci(docker-publish): limit workflow triggers in [#132](https://github.com/juicity/juicity/pull/132) by (@yqlbu)

### 特性

- 特性(cmd/version): 在 [#134](https://github.com/juicity/juicity/pull/134) 中添加版权和许可信息 by (@yqlbu)
- 特性(cmd): 在 [#123](https://github.com/juicity/juicity/pull/123) 中增强版本打印功能 by (@yqlbu)

### 其他

- 自动化(lint): 在 [#142](https://github.com/juicity/juicity/pull/142) 中升级 linter 版本 by (@sumire88)
- 杂项: 在 [#141](https://github.com/juicity/juicity/pull/141) 中将 softwind 和 quic-go 升级至 v0.40.1 by (@mzz2017)
- 自动化(docker-publish): 在 [#132](https://github.com/juicity/juicity/pull/132) 中限制工作流触发条件 by (@yqlbu)

**Full Changelog**: https://github.com/juicity/juicity/compare/v0.3.0...v0.4.0rc1

### New Contributors

- @sumire88 made their first contribution in #142

### v0.3.0 (Latest)

> Release date: 2023/09/02

### Bug Fixes

- fix: disable_outbound_udp443 not work and revert #103 in [#107](https://github.com/juicity/juicity/pull/107) by (@mzz2017)

### Others

- chore: fix IsGSOError judgement in [#109](https://github.com/juicity/juicity/pull/109) by (@mzz2017)

### 问题修复

- 修复: disable_outbound_udp443 不起作用并还原 #103 in [#107](https://github.com/juicity/juicity/pull/107) 由 (@mzz2017)

### 其他变更

- 杂项: 修复 IsGSOError 判断 in [#109](https://github.com/juicity/juicity/pull/109) 由 (@mzz2017)

**Full Changelog**: https://github.com/juicity/juicity/compare/v0.2.1...v0.3.0

### v0.2.1

> Release date: 2023/08/27

### Features

- optimize: use raw udp conn to solve quic in quic problem in [#103](https://github.com/juicity/juicity/pull/103) by (@mzz2017)

### Others

- chore: bump quic-go to v0.38.0 in [#101](https://github.com/juicity/juicity/pull/101) by (@mzz2017)
- chore: upgrade quic-go to v0.37.6 in [#100](https://github.com/juicity/juicity/pull/100) by (@mzz2017)

### 特性支持

- 优化: 使用原始的 UDP 连接来解决 QUIC in QUIC 的问题 in [#103](https://github.com/juicity/juicity/pull/103) by (@mzz2017)

### 其他变更

- 杂项: 升级 quic-go 至 v0.38.0 in [#101](https://github.com/juicity/juicity/pull/101) by (@mzz2017)
- 杂项: 升级 quic-go 至 v0.37.6 in [#100](https://github.com/juicity/juicity/pull/100) by (@mzz2017)

**Full Changelog**: https://github.com/juicity/juicity/compare/v0.2.0...v0.2.1

### v0.2.0

> Release date: 2023/08/19

### Features

- feat(server): support dial using socks5, http, vmess, vless... in [#92](https://github.com/juicity/juicity/pull/92) by (@mzz2017)
- feat: support port forwarding in [#86](https://github.com/juicity/juicity/pull/86) by (@mzz2017)

### Bug Fixes

- fix: fail to trigger auth timeout in some cases in [#93](https://github.com/juicity/juicity/pull/93) by (@mzz2017)
- fix: panic: unaligned 64-bit atomic operation in [#91](https://github.com/juicity/juicity/pull/91) by (@mzz2017)

### Others

- chore: bump quic-go to v0.37.5 in [#96](https://github.com/juicity/juicity/pull/96) by (@mzz2017)

### 特性支持

- 特性(服务器): 支持使用 socks5、http、vmess、vless 等进行拨号 in [#92](https://github.com/juicity/juicity/pull/92) by (@mzz2017)
- 特性: 支持端口转发 in [#86](https://github.com/juicity/juicity/pull/86) by (@mzz2017)

### 问题修复

- 修复: 某些情况下无法触发身份验证超时 in [#93](https://github.com/juicity/juicity/pull/93) by (@mzz2017)
- 修复: panic: 未对齐的 64 位原子操作 in [#91](https://github.com/juicity/juicity/pull/91) by (@mzz2017)

### 其他变更

- 杂项: 升级 quic-go 至 v0.37.5 in [#96](https://github.com/juicity/juicity/pull/96) by (@mzz2017)

**Full Changelog**: https://github.com/juicity/juicity/compare/v0.1.3...v0.2.0

### v0.1.3

> Release date: 2023/08/13

### Features

- optimize(generate-sharelink): judge whether really need to pin in [#80](https://github.com/juicity/juicity/pull/80) by (@mzz2017)
- feat(server): support cmd generate-sharelink in [#74](https://github.com/juicity/juicity/pull/74) by (@mzz2017)

### Others

- docs: add lang switch for spec.md in [#85](https://github.com/juicity/juicity/pull/85) by (@mzz2017)
- docs(spec): add English version specification in [#84](https://github.com/juicity/juicity/pull/84) by (@mzz2017)
- docs: add protocol spec in [#83](https://github.com/juicity/juicity/pull/83) by (@mzz2017)
- chore: bump quic-go to v0.37.4 to support go1.21 in [#81](https://github.com/juicity/juicity/pull/81) by (@mzz2017)
- chore(pr_template): update pr_template section headers in [#79](https://github.com/juicity/juicity/pull/79) by (@yqlbu)
- docs(readme): add self-signed certs to goals in [#78](https://github.com/juicity/juicity/pull/78) by (@yqlbu)
- chore/refactor: rework issue_templates in [#77](https://github.com/juicity/juicity/pull/77) by (@yqlbu)
- ci/optimize: docker refinement in [#75](https://github.com/juicity/juicity/pull/75) by (@yqlbu)
- ci: add Dockerfile in [#73](https://github.com/juicity/juicity/pull/73) by (@EkkoG)
- chore: upgrade quic-go to v0.37.3 in [#70](https://github.com/juicity/juicity/pull/70) by (@mzz2017)

### 特性支持

- 优化(生成共享链接): 判断是否真的需要固定 in [#80](https://github.com/juicity/juicity/pull/80) by (@mzz2017)
- 特性(服务器): 支持命令生成共享链接 in [#74](https://github.com/juicity/juicity/pull/74) by (@mzz2017)

### 其他变更

- 文档: 为 spec.md 添加语言切换 in [#85](https://github.com/juicity/juicity/pull/85) by (@mzz2017)
- 文档(spec): 添加英文版本规范 in [#84](https://github.com/juicity/juicity/pull/84) by (@mzz2017)
- 文档: 添加协议规范 in [#83](https://github.com/juicity/juicity/pull/83) by (@mzz2017)
- 杂项: 升级 quic-go 至 v0.37.4 以支持 go1.21 in [#81](https://github.com/juicity/juicity/pull/81) by (@mzz2017)
- 杂项(pr_template): 更新 pr_template 部分标题 in [#79](https://github.com/juicity/juicity/pull/79) by (@yqlbu)
- 文档(readme): 在目标中添加自签名证书 in [#78](https://github.com/juicity/juicity/pull/78) by (@yqlbu)
- 杂项/重构: 重新设计 issue_templates in [#77](https://github.com/juicity/juicity/pull/77) by (@yqlbu)
- 自动化(优化): Docker 优化 in [#75](https://github.com/juicity/juicity/pull/75) by (@yqlbu)
- 自动化: 在 [#73](https://github.com/juicity/juicity/pull/73) 中添加 Dockerfile by (@EkkoG)
- 杂项: 将 quic-go 升级至 v0.37.3 in [#70](https://github.com/juicity/juicity/pull/70) by (@mzz2017)

**Full Changelog**: https://github.com/juicity/juicity/compare/v0.1.2...v0.1.3

### New Contributors

- @EkkoG made their first contribution in #73

### v0.1.2 (Latest)

> Release date: 2023/08/05

### Features

- feat: support certificate pinning in [#61](https://github.com/juicity/juicity/pull/61) by (@mzz2017)
- feat(internal/relay): add support to inspect source for UDP traffic in [#57](https://github.com/juicity/juicity/pull/57) by (@yqlbu)
- feat(log): add exra cmd-flags for file logger in [#56](https://github.com/juicity/juicity/pull/56) by (@yqlbu)
- feat(sever): add support to inspect source for TCP traffic in [#53](https://github.com/juicity/juicity/pull/53) by (@yqlbu)

### Bug Fixes

- fix(server): support to dial domain in UDP (quic) in [#54](https://github.com/juicity/juicity/pull/54) by (@mzz2017)
- patch(server): fix golang lint warning in [#51](https://github.com/juicity/juicity/pull/51) by (@yqlbu)

### Others

- chore: upgrade softwind and quic(to 0.37.2) in [#67](https://github.com/juicity/juicity/pull/67) by (@mzz2017)
- docs(server): add UUID Generator section in [#64](https://github.com/juicity/juicity/pull/64) by (@yqlbu)
- ci: add linting workflow in [#52](https://github.com/juicity/juicity/pull/52) by (@yqlbu)

### 特性支持

- 特性: 支持证书固定 in [#61](https://github.com/juicity/juicity/pull/61) by (@mzz2017)
- 特性(内部/中继): 添加支持打印 UDP 来源 in [#57](https://github.com/juicity/juicity/pull/57) by (@yqlbu)
- 特性(日志): 为文件日志添加额外的命令标志 in [#56](https://github.com/juicity/juicity/pull/56) by (@yqlbu)
- 特性(服务器): 添加支持打印 TCP 来源 in [#53](https://github.com/juicity/juicity/pull/53) by (@yqlbu)

### 问题修复

- 修复(服务器): 支持在 UDP（quic）中拨号到域名 in [#54](https://github.com/juicity/juicity/pull/54) by (@mzz2017)
- 补丁(服务器): 修复 golang lint 报错 in [#51](https://github.com/juicity/juicity/pull/51) by (@yqlbu)

### 其他变更

- 杂项: 升级 softwind 和 quic (到 0.37.2) in [#67](https://github.com/juicity/juicity/pull/67) by (@mzz2017)
- 文档(服务器): 添加 UUID 生成器部分 in [#64](https://github.com/juicity/juicity/pull/64) by (@yqlbu)
- 自动化: 添加 linting 工作流程 in [#52](https://github.com/juicity/juicity/pull/52) by (@yqlbu)

**Full Changelog**: https://github.com/juicity/juicity/compare/v0.1.1...v0.1.2

### v0.1.1

> Release date: 2023/07/31

### Features

- feat(client): add protect_path support for android vpn in [#44](https://github.com/juicity/juicity/pull/44) by (@arm64v8a)
- optimize: upgrade quic-go to v0.37.0 to support GSO in [#40](https://github.com/juicity/juicity/pull/40) by (@mzz2017)
- feat/refactor(pkg/log): add log writer stream in [#39](https://github.com/juicity/juicity/pull/39) by (@yqlbu)
- optimize: little more faster in [#36](https://github.com/juicity/juicity/pull/36) by (@mzz2017)
- feat(server): support fwmark in [#18](https://github.com/juicity/juicity/pull/18) by (mzz2017)
- feat: support send_through in [#17](https://github.com/juicity/juicity/pull/17) by (mzz2017)

### Bug Fixes

- fix: should support hex fwmark in [#20](https://github.com/juicity/juicity/pull/20) by (mzz2017)
- hotfix: fix server instantiation in [#19](https://github.com/juicity/juicity/pull/19) by (@yqlbu)

### Others

- feat/refactor: add constant pkg in [#31](https://github.com/juicity/juicity/pull/31) by (@yqlbu)
- ci/hotfix(build): fix secret inputs missing issue in [#33](https://github.com/juicity/juicity/pull/33) by (@yqlbu)
- docs: add authentication docs of http/socks5 listening in [#46](https://github.com/juicity/juicity/pull/46) by (@mzz2017)
- chore/fix: upgrade softwind to support to set udp buffer size in [#45](https://github.com/juicity/juicity/pull/45) by (@mzz2017)
- chore(pr_template): fix typos in [#43](https://github.com/juicity/juicity/pull/43) by (@yqlbu)
- ci: add generate-changelogs workflow in [#37](https://github.com/juicity/juicity/pull/37) by (@yqlbu)
- ci(build,daily-build): demise post-actions stage in [#35](https://github.com/juicity/juicity/pull/35) by (@yqlbu)
- ci(build,pr-build,daily-build): enable check_run stages in [#32](https://github.com/juicity/juicity/pull/32) by (@yqlbu)
- chore: add example-{client,server}.json in [#29](https://github.com/juicity/juicity/pull/29) by (@mzz2017)
- chore/optimize(log): change log format to datetime in [#26](https://github.com/juicity/juicity/pull/26) by (@mzz2017)
- docs: refine README in [#25](https://github.com/juicity/juicity/pull/25) by (@mzz2017)
- docs(readme): add daed client as a new juicity client in [#21](https://github.com/juicity/juicity/pull/21) by (@mzz2017)
- refactor/patch: rework logger instantiation in [#16](https://github.com/juicity/juicity/pull/16) by (@yqlbu)
- ci/feature: add release builds in [#15](https://github.com/juicity/juicity/pull/15) by (@yqlbu)

### 特性支持

- 特性(client): 为安卓 vpn 添加 protect_path 支持 in [#44](https://github.com/juicity/juicity/pull/44) by (@arm64v8a)
- 优化: 升级 quic-go 到 v0.37.0 以支持 GSO in [#40](https://github.com/juicity/juicity/pull/40) by (@mzz2017)
- 特性/重构(pkg/log): 添加日志写入流 in [#39](https://github.com/juicity/juicity/pull/39) by (@yqlbu)
- 优化: 小幅提升速度 in [#36](https://github.com/juicity/juicity/pull/36) by (@mzz2017)
- 特性(server): 支持 fwmark in [#18](https://github.com/juicity/juicity/pull/18) by (mzz2017)
- 特性: 支持 send_through in [#17](https://github.com/juicity/juicity/pull/17) by (mzz2017)

### 问题修复

- 修复: 应当支持十六进制 fwmark in [#20](https://github.com/juicity/juicity/pull/20) by (mzz2017)
- 紧急修复: 修复服务器实例化问题 in [#19](https://github.com/juicity/juicity/pull/19) by (@yqlbu)

### 其他变更

- 特性/重构: 添加常量包 in [#31](https://github.com/juicity/juicity/pull/31) by (@yqlbu)
- 自动化/修复(build): 修复 secret 输入丢失问题 in [#33](https://github.com/juicity/juicity/pull/33) by (@yqlbu)
- 文档: 添加 http/socks5 监听的身份验证文档 in [#46](https://github.com/juicity/juicity/pull/46) by (@mzz2017)
- 杂项/修复: 升级 softwind 以支持设置 UDP 缓冲区大小 in [#45](https://github.com/juicity/juicity/pull/45) by (@mzz2017)
- 杂项(pr_template): 修复拼写错误 in [#43](https://github.com/juicity/juicity/pull/43) by (@yqlbu)
- 自动化: 添加 generate-changelogs 工作流程 in [#37](https://github.com/juicity/juicity/pull/37) by (@yqlbu)
- 自动化(build,daily-build): 移除 post-actions 阶段 in [#35](https://github.com/juicity/juicity/pull/35) by (@yqlbu)
- 自动化(build,pr-build,daily-build): 启用 check_run 阶段 in [#32](https://github.com/juicity/juicity/pull/32) by (@yqlbu)
- 杂项: 添加 example-{client,server}.json in [#29](https://github.com/juicity/juicity/pull/29) by (@mzz2017)
- 杂项/优化(log): 更改日志格式为日期时间 in [#26](https://github.com/juicity/juicity/pull/26) by (@mzz2017)
- 文档: 优化 README in [#25](https://github.com/juicity/juicity/pull/25) by (@mzz2017)
- 文档(readme): 将 daed client 添加为新的 juicity client in [#21](https://github.com/juicity/juicity/pull/21) by (@mzz2017)
- 重构/补丁: 重新设计日志记录实例化 in [#16](https://github.com/juicity/juicity/pull/16) by (@yqlbu)
- 自动化/特性: 添加发布构建 in [#15](https://github.com/juicity/juicity/pull/15) by (@yqlbu)

**Full Changelog**: https://github.com/juicity/juicity/compare/v0.1.0...v0.1.1

### New Contributors

- @arm64v8a made their first contribution in #44

### v0.1.0

> Release date: 2023/07/30

### Notes

> **Note**: initial release

Juicity is a quic-based proxy protocol. It has strong performance and is of great help to improve the network quality of the proxy. We have mature experience in proxy protocol, which can ensure that you avoid procedural and design problems as much as possible when using them. Have fun!

Juicity 是一个基于 quic 的代理协议，它有着强劲的性能，对网络质量较差的代理环境有较大的改善。我们在代理协议上具有成熟的经验，能保证您在使用时尽可能避免程序性和设计性的问题。最后，玩得开心！
