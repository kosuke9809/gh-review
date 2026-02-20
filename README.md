# gh-review

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

## 機能

- **PR一覧** — CI状態（✓/✗/●）、レビュー進捗、レビューステータスバッジ（`[NEW]`/`[UPD]`/`[DONE]`/`[CHG]`）を一覧表示
- **詳細表示** — PRのCI チェック・レビュー・コメントをまとめて確認
- **差分表示** — ファイル一覧とsyntax-highlightされた差分を分割表示
- **git worktree 統合** — `Enter` キー一発でPRブランチのworktreeを作成、`o` キーでエディタを起動
- **フィルタ切り替え** — `f` キーでReview Requested / Authored / All Open をAPIコールなしで切り替え
- **自動リフレッシュ** — 60秒ごとにバックグラウンドで更新

## 必要環境

- [gh CLI](https://cli.github.com/) がインストール済みで認証済み（`gh auth login`）
- Go 1.24以上（ソースからビルドする場合）

## インストール

```bash
gh extension install kosuke9809/gh-review
```

## 使い方

PR を確認したいリポジトリのディレクトリで実行します。

```bash
gh review
```

### キーバインド

| キー | 操作 |
|------|------|
| `1` / `2` / `3` | タブ切り替え（PRs / Detail / Diff） |
| `↑` / `↓` / `j` / `k` | PR選択を移動 |
| `Enter` | 選択したPRの git worktree を作成 |
| `o` | worktreeをVS Codeで開く（worktree作成済みの場合） |
| `d` | Diffタブに移動 |
| `f` | フィルタを切り替え |
| `r` | 手動リフレッシュ |
| `tab` | Diffタブでファイル一覧↔差分ビューの切り替え |
| `q` / `Ctrl+C` | 終了 |

### レビューステータスバッジ

| バッジ | 意味 |
|--------|------|
| `[NEW]` | まだレビューしていない |
| `[UPD]` | あなたのレビュー後にPRが更新された |
| `[DONE]` | あなたがApproveした（その後更新なし） |
| `[CHG]` | あなたがChanges Requestedした |

### フィルタ

| フィルタ | 表示対象 |
|----------|----------|
| Review Requested | あなたにレビューリクエストされたPR |
| Authored | あなたが作成したPR |
| All Open | すべてのオープンPR |

### git worktree について

`Enter` を押すと、選択したPRブランチを `<リポジトリルート>/.worktrees/pr-<番号>/` に worktree として展開します。
複数のPRを同時に開いて作業できます。worktreeが作成済みのPRには `[wt]` バッジが表示されます。

## ソースからビルド

```bash
git clone https://github.com/kosuke9809/gh-review
cd gh-review
go build -o gh-review .
gh extension install .
```

## ライセンス

MIT
