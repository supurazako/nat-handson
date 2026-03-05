# nat-handson

低レイヤーズさっぽろのNAT体験のためのrepoです

Docker上に「client / router / server」構成を作成し、routerコンテナ上でNAPTを実装します。

## 目的

NAPTのしくみを学ぶ

## やること

本ハンズオンでは以下を体験します

- NAPT（IP + Port変換）の実装
- conntrackによる状態管理の観察
- 複数クライアント通信の検証
- DNATによるポートフォワーディング

## 前提環境

以下のいずれかの環境が必要です

- Linux環境
- Docker / Docker Compose が利用可能な環境

## クイックスタート

```bash
git clone https://github.com/supurazako/nat-handson.git
cd nat-handson
docker compose up -d
```

詳細な手順は[instructions.md](docs/instructions.md)を参照してください。

## ディレクトリ構成

- docker-compose.yml
- docs/instructions.md
- docs/troubleshoot.md

## ハンズオン構成

- Step1: 基本NAPT
- Step2: 複数クライアント通信
- Step3: 状態破壊
- Step4: ルール削除（発展）
- Step5: DNAT（発展）

## ゴール

NATは「状態を持つアドレス変換装置」であると理解すること。
