# Taskalyst
The Catalyst for Task Management

![Go](https://img.shields.io/badge/Backend-Go-00ADD8?style=flat-square&logo=go)
![PostgreSQL](https://img.shields.io/badge/Database-PostgreSQL-336791?style=flat-square&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=white&style=for-the-badge)

## 1. 概要
Taskalystは学生向けのセルフホスト型タスク管理・タイムトラッキングツールです．
個々のタスクの消化だけでなく，日々の活動を24/7で監視し，統合的に記録・可視化することを目的としています．

## 2. 特徴
* プロジェクトベース: タイムトラッキング，タスクは全てプロジェクト単位で行うため，何に時間を費やしたかを明確にできる．
* 自動化: 講義や予定表の記録を自動化する．
* 可視化: スキルアップに費やした時間をGitHubのContributionのように可視化し，モチベーションを維持する．
* クラウド非依存: 全てのデータはSelf-hosted環境に置く．外部APIへの依存を排除し，永続的な利用を目指す．
* API: 外部システムとの連携用に，Personal Access Tokenを発行可能．蓄積されたログをJSON形式でエクスポートし自作のツールなどで可視化できる．

## 3. インストール
現在製作中です．

## 4. ロードマップ

### Phase 1: 認証基盤とデータ基盤の構築
- [x] ユーザー管理・認証システム
- [ ] 2階層カテゴリ・プロジェクト管理
- [ ] 外部連携用認証

### Phase 2: コアタスク管理とタイムトラッキング
- [ ] タスク管理(V1)
    - 小タスク
    - Markdownメモ
    - 動的な進捗率
- [ ] タイムトラッキング(V1)
    - タイマー開始/停止/手動入力機能
    - 基本的な実績一覧取得 API

### Phase 3: 時間割とカレンダー同期(CalDAV)
- [ ] 時間割
- [ ] CalDAVサーバー実装
    - WebDAVプロトコル対応
    - iCalendar形式のエンコード/デコード
    - 終日予定・繰り返し予定
- [ ] 単発予定管理
    - カレンダー画面用の予定登録・編集機能

### Phase 4: 可視化とフォーカス機能
- [ ] 実績可視化
    - スキルアップ時間のContribution Graph集計 API
    - カテゴリ別の統計
- [ ] ポモドーロ・タイマー
    - 作業/休憩サイクル設定
    - フォーカスモード中の通知
- [ ] データエクスポート
    - 蓄積データのJSON形式出力機能

### Phase 5: フロントエンド・モバイル展開
- [ ] Web フロントエンド
- [ ] ハイブリッドアプリ化
    - オフライン時のデータキャッシュ（参照用）
    - ローカル通知によるタスク期限/予定の通知
- [ ] インフラ展開
