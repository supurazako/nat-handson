# NAT Hands-on Instructions

この手順では Docker 上に `client1 / client2 / router / server` を起動し、`router` で NAPT を設定して観察します。

## 前提

- Docker Desktop (Mac/Windows) または Docker Engine + Compose が使えること
- プロジェクトルートでコマンドを実行すること

## Step0: 今回使うコマンド

Step1で使うコマンドを先にまとめます。実行自体は Step1 の順序に従ってください。

### `ip`

役割

- Linux のネットワーク情報（IPアドレス、インターフェース、ルーティング）を確認・変更するコマンドです。

このハンズオンで使うコマンド

- `ip addr`（詳細: Step1-3, Step1-4）
  - IF名とIPを確認します。`client1` では `192.168.10.2/24`、`client2` では `192.168.10.3/24` が付いたIFを探し、`router` では `192.168.10.1/24` (LAN) と `172.31.0.1/24` (WAN) のIFを特定します。
- `ip route`（詳細: Step1-3）
  - 既定経路（default route）を確認します。
- `ip route del default` / `ip route add default via 192.168.10.1 dev <IF>`（詳細: Step1-3）
  - `client1` / `client2` の通信を Docker 既定gateway ではなく `router` に流すために変更します。

読み方のポイント

- `default via ... dev ...` が外向き通信の出口です。
- Alpine の BusyBox `ip` は `-br` 非対応なので `ip addr` を使います。

### `iptables`

役割

- Linux カーネルのパケット処理ルール（NAT/フィルタ）を設定・観察するコマンドです。

このハンズオンで使うコマンド

- `iptables -t nat -A POSTROUTING -s 192.168.10.0/24 -o <WAN_IF> -j MASQUERADE`（詳細: Step1-4）
  - LAN 側アドレスを WAN IF のアドレスに変換する NAPT 設定です。
- `iptables -A FORWARD -i <LAN_IF> -o <WAN_IF> -j ACCEPT`（詳細: Step1-4）
  - 行きの通信（LAN->WAN）を許可します。
- `iptables -A FORWARD -i <WAN_IF> -o <LAN_IF> -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT`（詳細: Step1-4）
  - 戻り通信（WAN->LAN）を接続状態ベースで許可します。
- `iptables -t nat -L -n -v` / `iptables -L -n -v`（詳細: Step1-6）
  - ルールのヒット回数（`pkts/bytes`）を確認します。

読み方のポイント

- `MASQUERADE` の `pkts/bytes` が増える = NAT ルールに実トラフィックが当たっている。
- `FORWARD` 2本のカウンタが増える = 行き/戻りの転送が成立している。

### `conntrack`

役割

- Linux の接続追跡テーブル（stateful NAT/FW の状態管理）を表示するコマンドです。

このハンズオンで使うコマンド

- `conntrack -L`（詳細: Step1-6）
  - 現在追跡中の通信フローを一覧表示します。

読み方のポイント

- `src` / `dst`: 通信元/宛先IP
- `sport` / `dport`: 通信元/宛先ポート
- `ESTABLISHED` / `TIME_WAIT` など: 接続状態
- 成功判定では `src=192.168.10.2 dst=172.31.0.2 dport=80` 系が見えるかを確認します。

### `docker compose`

- `docker compose up -d`（Step1-1）: 構成を起動します。
- `docker compose ps`（Step1-1）: 起動状態を確認します。
- `docker compose exec router sh`（Step1-2）/ `docker compose exec client1 sh`（Step1-3）/ `docker compose exec client2 sh`（Step2-2）: 各コンテナに入ります。

### `apk`

- `apk` は Alpine Linux のパッケージ管理コマンドです。
- このハンズオンでは不足ツールをその場で導入します（Step1-2, Step1-5）。

### `curl`

- `curl` は HTTP リクエストを送る確認用コマンドです。
- `curl http://172.31.0.2`（Step1-5, Step2-3）で `client1/client2 -> server` の疎通を確認します。

## Step1: 基本NAPT

### 1. 起動

```bash
docker compose up -d
docker compose ps
```

`nat-client1` / `nat-client2` / `nat-router` / `nat-server` が `Up` になっていることを確認します。

### 2. router に入って必要パッケージをインストール

```bash
docker compose exec router sh
apk add --no-cache iptables iproute2 conntrack-tools tcpdump
```

このシェルは後続の確認で使うため、開いたままでも問題ありません。

### 3. client1 の default route を 192.168.10.1 に変更

別ターミナルで `client1` に入ります。

```bash
docker compose exec client1 sh
ip addr
ip route
```

`192.168.10.2/24` が付いている IF を確認し、その IF を使って default route を設定します。

[ip addr 実行例](images/ip-addr.png)

以下は IF 名を `eth0` と仮定した例です。

```bash
ip route del default
ip route add default via 192.168.10.1 dev eth0
ip route
```

### 4. router で MASQUERADE と FORWARD を設定

`router` シェルで LAN/WAN の IF 名を確認します。

```bash
ip addr
```

`192.168.10.1/24` が付いている IF を `LAN_IF`、`172.31.0.1/24` が付いている IF を `WAN_IF` として、以下を実行します。
おそらく `eth0` が LAN、`eth1` が WAN になると思いますが、環境によって異なる可能性があるため、確認してから実行してください。
`<LAN_IF>` と `<WAN_IF>` はそれぞれの IF 名に置き換えてください。(例: `eth0` / `eth1`)

Alpine の BusyBox `ip` は `-br` オプション非対応のため、`ip addr` を使います。

```bash
iptables -t nat -A POSTROUTING -s 192.168.10.0/24 -o <WAN_IF> -j MASQUERADE
iptables -A FORWARD -i <LAN_IF> -o <WAN_IF> -j ACCEPT
iptables -A FORWARD -i <WAN_IF> -o <LAN_IF> -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
```

### 5. client1 から server へアクセス確認

`client1` シェルで実行します。
Alpine の `client1` には `curl` が入っていない場合があるため、このステップで導入します。

```bash
apk add --no-cache curl
curl http://172.31.0.2
```

`Welcome to nginx!` のHTMLが返れば成功です。

### 6. router で conntrack と NAT テーブル確認

`router` シェルで実行します。

```bash
conntrack -L
iptables -t nat -L -n -v
iptables -L -n -v
```

`conntrack` にフローが作成され、`POSTROUTING` と `FORWARD` のカウンタが増えていることを確認します。

### 7. 成功判定（Before / After 比較）

Step1.5 を実行する前後で、次の観点を比較すると判断しやすくなります。

| 観点 | 未成功時（通信前/未ヒット） | 成功時（通信後/ヒット済み） | 判定ポイント |
| --- | --- | --- | --- |
| `conntrack -L` | 0件、または `dst=172.31.0.2` が見えない | `src=192.168.10.2 dst=172.31.0.2 dport=80` 系が見える | 目的通信のセッションがあるか |
| `iptables -t nat -L -n -v` | `POSTROUTING` の `MASQUERADE` が `0 0` のまま | `MASQUERADE` の `pkts/bytes` が 0 より大きい | NATルールに実トラフィックが当たったか |
| `iptables -L -n -v` | `FORWARD` の2ルールが `0 0` のまま | `LAN->WAN` と `WAN->LAN (ESTABLISHED,RELATED)` のカウンタが増える | 転送の往復が成立しているか |

補足:
- `127.0.0.11` 向けの Docker DNS 通信や外部IP向け通信が混在していても問題ありません。
- 判定は必ず「`192.168.10.2 -> 172.31.0.2:80` の通信が見えているか」で行ってください。

### 8. NAPTでどのIP:Portがどこに配送されたかを確認

このStep1で確認している変換は次のイメージです。

- 送信元（LAN側・変換前）: `192.168.10.2:<client_ephemeral_port>`
- 宛先（server）: `172.31.0.2:80`
- 送信元（WAN側・変換後）: `172.31.0.1:<translated_port>`
  `MASQUERADE` のため、送信元IPは `router` の WAN 側IPに変換されます。

確認用に、`client1` で `curl` の送信元ポートを表示します。

```bash
curl -s -o /dev/null -w 'client_local=%{local_ip}:%{local_port} -> server=%{remote_ip}:%{remote_port}\n' http://172.31.0.2
```

続けて `router` で同じフローを確認します。

```bash
conntrack -L | grep 'src=192.168.10.2' | grep 'dst=172.31.0.2' | grep 'dport=80'
```

`conntrack` の1行は、前半が「client1 -> server」、後半が「server -> router(WAN)」の対応を表します。
以下は読み方の例です（ポート番号は毎回変わります）。

```text
tcp ... TIME_WAIT src=192.168.10.2 dst=172.31.0.2 sport=<client_port> dport=80 src=172.31.0.2 dst=172.31.0.1 sport=80 dport=<translated_port> [ASSURED] ...
```

具体値の例:

```text
tcp 6 112 TIME_WAIT src=192.168.10.2 dst=172.31.0.2 sport=38624 dport=80 src=172.31.0.2 dst=172.31.0.1 sport=80 dport=38624 [ASSURED] mark=0 use=1
```

- 前半（original方向）: `src=192.168.10.2:<client_port> -> dst=172.31.0.2:80`
- 後半（reply方向）: `src=172.31.0.2:80 -> dst=172.31.0.1:<translated_port>`
- `sport` / `dport` は送信元/宛先ポートです。
- `TIME_WAIT` や `ESTABLISHED` はTCPの状態、`[ASSURED]` は双方向通信が成立した状態です。
- `<client_port>` と `<translated_port>` は例です。実行ごとに別の値になります。

なぜ戻り先が `router` の WAN 側になるのか:
- `client1 -> server` の送信時、`router` は `MASQUERADE` により送信元を `172.31.0.1:<translated_port>` に変換します。
- そのため `server` からは通信相手が `172.31.0.1:<translated_port>` に見えます。
- `server` の返信先は見えている相手になるため、戻りパケットの宛先は `dst=172.31.0.1:<translated_port>` になります。
- 受け取った `router` は `conntrack` の状態を使って逆変換し、`192.168.10.2:<client_port>` へ転送します。
- 重要なのは後半の `dst` が `router` WAN 側IPになっている点です（ポート値自体は毎回変わります）。

`conntrack` に対象フローが出て、`iptables -t nat -L -n -v` の `MASQUERADE` カウンタが増えていれば、NAPTによる配送が機能しています。

やっていることのイメージ画像

- [NAPT send](images/napt-send.png)
- [NAPT reply](images/napt-reply.png)
- [NAPT table](images/napt-table.png)

## Step2: 複数クライアント通信

このStepでは `client1` と `client2` の両方から `server` に通信し、`router` で別々のフローとして追跡されることを確認します。

### 1. 起動状態を確認

```bash
docker compose up -d
docker compose ps
```

`nat-client1` / `nat-client2` / `nat-router` / `nat-server` が `Up` であることを確認します。

### 2. client1 / client2 の default route を router に向ける

`client1`:

```bash
docker compose exec client1 sh
ip route del default
ip route add default via 192.168.10.1 dev eth0
ip route
```

`client2`:

```bash
docker compose exec client2 sh
ip route del default
ip route add default via 192.168.10.1 dev eth0
ip route
```

どちらも `default via 192.168.10.1 dev eth0` になっていることを確認します。

### 3. client1 / client2 から server へアクセス

`client1`:

```bash
docker compose exec client1 sh -c "apk add --no-cache curl && curl -s http://172.31.0.2 >/dev/null && echo client1_ok"
```

`client2`:

```bash
docker compose exec client2 sh -c "apk add --no-cache curl && curl -s http://172.31.0.2 >/dev/null && echo client2_ok"
```

### 4. router で複数フローを観察

```bash
docker compose exec router conntrack -L
docker compose exec router iptables -t nat -L -n -v
docker compose exec router iptables -L -n -v
```

`conntrack` では次の2種類のフローが見えることを確認します。

- `src=192.168.10.2 ... dst=172.31.0.2 dport=80`（client1）
- `src=192.168.10.3 ... dst=172.31.0.2 dport=80`（client2）

### 5. Step2 の成功判定

- 2つの送信元IP（`192.168.10.2`, `192.168.10.3`）由来のエントリが `conntrack` に出る
- `MASQUERADE` と `FORWARD` のカウンタが増える
- 同じ宛先でも `src IP + src port` の組ごとに別セッションとして管理される

## Step3: 状態破壊

準備中。

## Step4: DNAT（挑戦）

準備中。
