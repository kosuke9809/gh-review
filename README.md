# gh-review【WIP】

GitHub のPRレビューをターミナルから一元管理する `gh` CLI 拡張機能です。

複数のオープンPRのCI状態・レビュー進捗・差分をキーボードだけで高速に確認し、git worktree を使ってPRブランチへ即座に切り替えられます。

```
gh-review — myorg/backend                [1:PRs] [2:Detail] [3:Diff]   [f] All Open
────────────────────────────────────────────────────────────────────────────────────
  #142  Fix authentication bug            CI:✓  Review:1/2  [UPD]  [wt]
  #89   Add dark mode support             CI:✗  Review:0/2  [NEW]
▶ #203  Refactor database layer           CI:✓  Review:2/2  [DONE] [wt]
  #15   Update terraform config           CI:●  Review:1/1  [CHG]

[Enter]worktree  [d]iff  [o]open  [f]filter  [r]efresh  [q]uit         Last sync: 5s ago
```

## インストール

```bash
gh extension install kosuke9809/gh-review
```
