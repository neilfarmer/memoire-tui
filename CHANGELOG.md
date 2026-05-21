# Changelog

## [0.4.0](https://github.com/neilfarmer/memoire-tui/compare/v0.3.0...v0.4.0) (2026-05-21)


### Features

* **assistant:** always show convo cursor + ctrl+k to jump to convos pane ([c45f61f](https://github.com/neilfarmer/memoire-tui/commit/c45f61f3b8776e30ea9f23fbecc8d609d27b8372))
* **assistant:** UI overhaul + fix dead ctrl+m model toggle ([#19](https://github.com/neilfarmer/memoire-tui/issues/19)) ([d9c2b09](https://github.com/neilfarmer/memoire-tui/commit/d9c2b09d3fcf43898db702b525d37ba6843d1a37))
* **tui:** ANSI Shadow memoire splash for boot + dashboard loading ([#18](https://github.com/neilfarmer/memoire-tui/issues/18)) ([d205174](https://github.com/neilfarmer/memoire-tui/commit/d2051748a6001533b7dbcdef2c9d065ece536a20))
* **tui:** wrap the entire app in a double-line outer border ([341e596](https://github.com/neilfarmer/memoire-tui/commit/341e5962c7f6adb3bddca0705cb611bd380aef1b))


### Bug Fixes

* **assistant:** correct rightWidth math so the body fits content area ([0b0f920](https://github.com/neilfarmer/memoire-tui/commit/0b0f92070a7f4aeb0a210b19cf11c151780df020))
* **assistant:** refocus input on re-entry; bind ctrl+y as global model toggle ([8e36c13](https://github.com/neilfarmer/memoire-tui/commit/8e36c1380f42da723871eb9ad83d0d902eca6046))
* **assistant:** restore input box visibility ([d61aee6](https://github.com/neilfarmer/memoire-tui/commit/d61aee6cb701b2600dac0ccf895243c64c1effab))
* **assistant:** stop ANSI codes leaking in cursor-highlighted convo row ([a9d552e](https://github.com/neilfarmer/memoire-tui/commit/a9d552eaf2b54983a3d1605059b0af4aefdf88f1))
* **habits:** window the rendered list so the cursor stays visible ([#16](https://github.com/neilfarmer/memoire-tui/issues/16)) ([d984f29](https://github.com/neilfarmer/memoire-tui/commit/d984f292568854eacc95748bfb8a039b2cb041bf))
* **tui:** capture delete target on 'd', not after table moves cursor ([#13](https://github.com/neilfarmer/memoire-tui/issues/13)) ([5c83fa7](https://github.com/neilfarmer/memoire-tui/commit/5c83fa77d6e3a056b80f2710bb482cc8b140301d))
* **tui:** delete the selected task, not the row above/below ([9817bb1](https://github.com/neilfarmer/memoire-tui/commit/9817bb1c5db3487294ba977901192fee8989cf85))
* **tui:** delete the selected task, not the row above/below ([246bf15](https://github.com/neilfarmer/memoire-tui/commit/246bf15c476b386496646f448f4d8f6f3406773d))
* **tui:** esc in assistant returns focus to sidebar instead of being swallowed ([#17](https://github.com/neilfarmer/memoire-tui/issues/17)) ([a7175ff](https://github.com/neilfarmer/memoire-tui/commit/a7175ffcc31f33630589f76ce529e20cf6875ce3))

## [0.3.0](https://github.com/neilfarmer/memoire-tui/compare/v0.2.0...v0.3.0) (2026-05-10)


### Features

* standardize keymap, add command palette, esc drill-up ([44cc256](https://github.com/neilfarmer/memoire-tui/commit/44cc2565148d3dbf7271625f7ad7af4e6349e4ed))
* standardize keymap, add command palette, esc drill-up ([d59b7a5](https://github.com/neilfarmer/memoire-tui/commit/d59b7a595bdc2b3b06b7c5a9ec6c7e7cdbbbd072))
* **ui:** fork bubbles/table for per-row stripe hook ([3ef5f96](https://github.com/neilfarmer/memoire-tui/commit/3ef5f96cd58065492405bc852d91dd5d3696a9e4))
* **ui:** outer frame, thicker borders, row stripes ([a688bb0](https://github.com/neilfarmer/memoire-tui/commit/a688bb066b03fff0111af56361d5374bd74f12dc))


### Bug Fixes

* arrow keys nav screens; cursor uses j/k ([59e72da](https://github.com/neilfarmer/memoire-tui/commit/59e72da9e06a7e2c94c49234ffc477a8fc29976a))
* arrow-only navigation; remove ctrl+n / j-k ([cd31202](https://github.com/neilfarmer/memoire-tui/commit/cd312021e8b0d0978fdc3aee5e948573e4af9448))
* **forms:** enter inserts newline in body textareas ([c01c4b3](https://github.com/neilfarmer/memoire-tui/commit/c01c4b3d89540e823bd4ca2649a7f51298443b3c))
* **nav:** sidebar always wins arrows / esc / : when focused ([200b73c](https://github.com/neilfarmer/memoire-tui/commit/200b73c60730a5eea9f9fd261f7d9d8e0ae5648d))
* **notes:** folders-first browse, drop F/X/f shortcuts ([3049896](https://github.com/neilfarmer/memoire-tui/commit/3049896152fc53eb3c64628be7ed107d2f9cc2a5))
* scope ? help overlay + drop phantom Nutrition screen ([89dd864](https://github.com/neilfarmer/memoire-tui/commit/89dd86426ff8e0fc8667e9ed5b2cac54eb003faa))
* scope ? help overlay + drop phantom Nutrition screen ([80a018f](https://github.com/neilfarmer/memoire-tui/commit/80a018fb8c69a4945416d0b599b2d1cd3b433d43))
* **ui:** drop Cell+Header padding so rows do not wrap ([dabf3ba](https://github.com/neilfarmer/memoire-tui/commit/dabf3bacfd2ba8dc3559409cdd70760d658c71c3))
* **ui:** drop oversized chrome — outer frame + thick borders broke layout ([2d16f90](https://github.com/neilfarmer/memoire-tui/commit/2d16f909a618f54fcdd36d30f75d9f8a672dd399))
* **ui:** full-row alternating background; drop leading icon column ([fed6315](https://github.com/neilfarmer/memoire-tui/commit/fed6315fbedf70b07264795dd7b23d747c1d7985))
* **ui:** revert broken Stripe — bubbles/table truncates ANSI cells ([29b8c3d](https://github.com/neilfarmer/memoire-tui/commit/29b8c3d555fe65ba25c9d98ff0553ca5e4d2d74b))
* **ui:** stripe only odd rows so alternation is visible ([b131f66](https://github.com/neilfarmer/memoire-tui/commit/b131f6619e7ec6715f5e1204b56b148fc2aefe5e))

## [0.2.0](https://github.com/neilfarmer/memoire-tui/compare/v0.1.0...v0.2.0) (2026-05-10)


### Features

* import memoire TUI ([30790be](https://github.com/neilfarmer/memoire-tui/commit/30790be05402b866bab259153121a2412b693d77))
* import memoire TUI Go client ([bb54706](https://github.com/neilfarmer/memoire-tui/commit/bb54706075ae65497de19120b26e5a023599b415))


### Bug Fixes

* **ci:** make pipeline pass ([bc96acd](https://github.com/neilfarmer/memoire-tui/commit/bc96acda0d904e15144f478a9d44caca30104154))
