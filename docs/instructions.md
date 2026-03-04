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
- `ip route del default` / `ip route add default via 192.168.10.1`（詳細: Step1-3）
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
ip route add default via 192.168.10.1
ip route
```

実行結果サンプル（抜粋）:

```text
default via 192.168.10.1 dev <LAN_IF>
192.168.10.0/24 dev eth0 scope link src 192.168.10.2
```

判定ポイント: `default via 192.168.10.1` になっていればOKです（`dev` 名は環境により異なります）。

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

実行結果サンプル（抜粋）:

```text
<!DOCTYPE html>
<title>Welcome to nginx!</title>
```

判定ポイント: `Welcome to nginx!` が含まれていればOKです。

### 6. router で conntrack と NAT テーブル確認

`router` シェルで実行します。

```bash
conntrack -L
iptables -t nat -L -n -v
iptables -L -n -v
```

`conntrack` にフローが作成され、`POSTROUTING` と `FORWARD` のカウンタが増えていることを確認します。

実行結果サンプル（抜粋）:

```text
tcp ... src=192.168.10.2 dst=172.31.0.2 sport=<client_port> dport=80 ...
...
POSTROUTING ...
<N> <BYTES> MASQUERADE ... 192.168.10.0/24 ...
FORWARD ...
<N> <BYTES> ACCEPT ... eth0 eth1 ...
```

判定ポイント:
- `conntrack` に `src=192.168.10.2` と `dst=172.31.0.2 dport=80` を含む行がある
- `conntrack` に `dst=172.31.0.1`（router WAN 側）または `reply_dst=172.31.0.1` 系の表示が見える
- `MASQUERADE` / `FORWARD` の `pkts/bytes` が 0 より増えている  
注記: `<client_port>` と `<N>` は実行ごとに変わります。

### 7. 成功判定（Before / After 比較）

Step1.5 を実行する前後で、次の観点を比較すると判断しやすくなります。

| 観点 | 未成功時（通信前/未ヒット） | 成功時（通信後/ヒット済み） | 判定ポイント |
| --- | --- | --- | --- |
| `conntrack -L` | 0件、または `dst=172.31.0.2` が見えない | `src=192.168.10.2 dst=172.31.0.2 dport=80` に加えて `dst=172.31.0.1`（または `reply_dst=...`）が見える | 変換前/変換後の対応が見えているか |
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

必要に応じて、NAT後（WAN側）の行も確認します。

```bash
conntrack -L | grep '172.31.0.1'
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

このStepのゴール

- 1つのグローバルIP（`router` WAN: `172.31.0.1`）で、複数クライアントを同時に外部へ出せる理由を理解する。

前提

- Step2 を単独で開始する場合でも、先に Step1-2（router の必要パッケージ導入）と Step1-4（MASQUERADE/FORWARD 設定）を実施してください。
- これらが未実施だと、`client1` / `client2` からの通信が成立しません。

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
ip route add default via 192.168.10.1
ip route
```

`client2`:

```bash
docker compose exec client2 sh
ip route del default
ip route add default via 192.168.10.1
ip route
```

どちらも `default via 192.168.10.1` になっていることを確認します（`dev` 名は環境により異なります）。

### 3. client1 / client2 から server へアクセス

`client1`:

```bash
curl -s http://172.31.0.2 >/dev/null && echo client1_ok
```

`client2`:

```bash
curl -s http://172.31.0.2 >/dev/null && echo client2_ok
```

実行結果サンプル（抜粋）:

```text
client1_ok
client2_ok
```

判定ポイント: `client1_ok` と `client2_ok` の両方が表示されればOKです。

### 4. router で複数フローを観察

router シェルで実行します。

```bash
conntrack -L
iptables -t nat -L -n -v
iptables -L -n -v
```

`conntrack` では次の2種類のフローが見えることを確認します。

- `src=192.168.10.2 ... dst=172.31.0.2 dport=80`（client1）
- `src=192.168.10.3 ... dst=172.31.0.2 dport=80`（client2）

実行結果サンプル（抜粋）:

```text
tcp ... src=192.168.10.2 dst=172.31.0.2 sport=<c1_port> dport=80 ...
tcp ... src=192.168.10.3 dst=172.31.0.2 sport=<c2_port> dport=80 ...
...
<N1> <BYTES1> MASQUERADE ... 192.168.10.0/24 ...
<N2> <BYTES2> ACCEPT ... eth0 eth1 ...
```

判定ポイント

- `192.168.10.2` と `192.168.10.3` の両方のフローが見える
- `dst=172.31.0.2 dport=80`（変換前）に加えて、`dst=172.31.0.1` または `reply_dst=172.31.0.1` 系（変換後）が見える
- `MASQUERADE` / `FORWARD` のカウンタが増える

注記: `<c1_port>`, `<c2_port>`, `<N1>`, `<N2>` は実行ごとに変わります。

### 5. Step2 の成功判定

- 観察A: `conntrack` に 2つの送信元IP（`192.168.10.2`, `192.168.10.3`）由来のエントリが出る
  理由対応: 別クライアントの通信が別セッションとして管理されている。
- 観察A-補足: `dst=172.31.0.2 dport=80` と `dst=172.31.0.1`（または `reply_dst=...`）の両方が見える
  理由対応: 変換前（LAN側）と変換後（WAN側）の対応が追跡できている。
- 観察B: `MASQUERADE` と `FORWARD` のカウンタが増える
  理由対応: 同時通信が実際に NAT/転送 された。
- 観察C: `client1_ok` と `client2_ok` の両方が表示される
  理由対応: 単一のWAN IPでも、2クライアントの通信を成立させられている。
- 上の「実行結果サンプル（抜粋）」と同じ観点を満たしていれば成功

### 6. なぜ1つのグローバルIPで同時通信できるのか

- `MASQUERADE` により、外側（WAN側）へ出るときの送信元IPはどちらも `172.31.0.1` になります。
- ただし送信元ポートは通信ごとに異なるため、`src IP + src port + dst IP + dst port` の組でフローを識別できます。
- `conntrack` はこの対応表（変換前/変換後）を保持しているため、戻り通信を正しい `client1` / `client2` に戻せます。
- つまり「IPを共有し、ポートと状態管理で多重化している」のが NAPT のポイントです。

注記

- ポート番号や `pkts/bytes` カウンタの具体値は実行ごとに変わります。値の一致ではなく、上記の構造で判定してください。

## Step3: 状態破壊

### 1. 目的

このStepのゴール

- NAPT が `conntrack` の状態に依存していることを体感する。
- `conntrack -F` で状態を消しても、通信再実行で状態が再生成されることを確認する。

### 2. 前提

- Step1 で `MASQUERADE/FORWARD` 設定済みであること
- Step2 で `client1` / `client2` の通信が成功していること

### 3. Flush前のエントリを作成して確認

まず `client1` / `client2` から通信を発生させ、Flush前にエントリがある状態を作ります。

`client1`:

```bash
curl -s http://172.31.0.2 >/dev/null && echo client1_prewarm_ok
```

`client2`:

```bash
curl -s http://172.31.0.2 >/dev/null && echo client2_prewarm_ok
```

続けて `router` シェルで確認します。

```bash
conntrack -L
```

実行結果サンプル（抜粋）:

```text
client1_prewarm_ok
client2_prewarm_ok
tcp ... src=192.168.10.2 dst=172.31.0.2 ...
tcp ... src=192.168.10.3 dst=172.31.0.2 ...
... dst=172.31.0.1 ...   # または reply_dst=172.31.0.1 系
```

判定ポイント

- `client1_prewarm_ok` / `client2_prewarm_ok` が出る
- `conntrack` に `src=192.168.10.2` と `src=192.168.10.3` が見える
- `dst=172.31.0.2 dport=80`（変換前）と `dst=172.31.0.1` または `reply_dst=...`（変換後）が見える

### 4. conntrack を Flush

`router` シェルで実行します。

```bash
conntrack -F
conntrack -L
```

実行結果サンプル（抜粋）:

```text
conntrack v1.4.8 (conntrack-tools): 0 flow entries have been shown.
```

判定ポイント

- Flush後に対象エントリが減る/消える

### 5. client1 / client2 から再通信

`client1`:

```bash
curl -s http://172.31.0.2 >/dev/null && echo client1_reconnect_ok
```

`client2`:

```bash
curl -s http://172.31.0.2 >/dev/null && echo client2_reconnect_ok
```

実行結果サンプル（抜粋）:

```text
client1_reconnect_ok
client2_reconnect_ok
```

### 6. router で conntrack 再生成を確認

```bash
conntrack -L
```

必要に応じて、変換後（WAN側）の表示も確認します。

```bash
conntrack -L | grep '172.31.0.1'
```

実行結果サンプル（抜粋）:

```text
tcp ... src=192.168.10.2 dst=172.31.0.2 sport=<c1_port> dport=80 ...
tcp ... src=192.168.10.3 dst=172.31.0.2 sport=<c2_port> dport=80 ...
... dst=172.31.0.1 ...   # または reply_dst=172.31.0.1 系
```

### 7. Step3 の成功判定

- Flush前に `src=192.168.10.2` / `src=192.168.10.3` のエントリが存在する
- Flush後に一度エントリが減る/消える
- 再通信後に `src=192.168.10.2` と `src=192.168.10.3` のエントリが再生成される
- `dst=172.31.0.2 dport=80`（変換前）と `dst=172.31.0.1` または `reply_dst=...`（変換後）が再び観察できる

### 8. なぜそうなるか

- NAT 変換ルール（`MASQUERADE/FORWARD`）は `iptables` に残っているため有効です。
- ただし、どの通信をどこへ戻すかの状態は `conntrack` が保持しています。
- `conntrack -F` で既存状態を消すと、進行中セッションは切れたり再接続になったりします。
- 新しく通信を送ると `conntrack` が状態を再学習し、エントリが再生成されます。

### 9. 注意

- 出力の形式・件数・ポート番号は環境で変わります。値の一致ではなく構造で判定してください。
- 短命なフローはすぐ消えることがあるため、通信直後に確認してください。
- Flush により既存接続が切れる/再接続になることがあります。

## Step4: ルール削除（発展）

### 1. 目的

このStepのゴール:
- NAT が成立するには、`conntrack` 状態だけでなく `iptables` の変換/転送ルールが必要だと体感する。

### 2. 前提

- Step1 の `MASQUERADE/FORWARD` 設定が完了していること
- `client1` / `client2` から `server` への疎通ができていること

### 3. 現状ルールを表示（保存用）

`router` シェルで実行します。

```bash
ip addr
iptables -t nat -S
iptables -S
```

`192.168.10.1/24` が付いた IF を `LAN_IF`、`172.31.0.1/24` が付いた IF を `WAN_IF` として控えてください。

### 4. 事前疎通確認

`client1`:

```bash
curl -s http://172.31.0.2 >/dev/null && echo client1_precheck_ok
```

`client2`:

```bash
curl -s http://172.31.0.2 >/dev/null && echo client2_precheck_ok
```

### 5. ルール削除（router）

`router` シェルで、以下3ルールを削除します。

- `POSTROUTING` の `MASQUERADE`
- `FORWARD` の `LAN -> WAN` 許可
- `FORWARD` の `WAN -> LAN` `ESTABLISHED,RELATED` 許可

削除例（`<LAN_IF>` / `<WAN_IF>` は自分の環境に合わせる）:

```bash
iptables -t nat -D POSTROUTING -s 192.168.10.0/24 -o <WAN_IF> -j MASQUERADE
iptables -D FORWARD -i <LAN_IF> -o <WAN_IF> -j ACCEPT
iptables -D FORWARD -i <WAN_IF> -o <LAN_IF> -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
```

注意:
- 環境差があるため、必ず `iptables -t nat -S` と `iptables -S` の出力を見て削除対象を確認してください。
- 削除前に復旧コマンドを控えておいてください。

### 6. 疎通失敗確認

`client1`:

```bash
curl http://172.31.0.2
```

`client2`:

```bash
curl http://172.31.0.2
```

失敗（タイムアウトや接続エラー）することを確認します。

### 7. ルール復旧（router）

`router` シェルで、削除した3ルールを再追加します。

```bash
iptables -t nat -A POSTROUTING -s 192.168.10.0/24 -o <WAN_IF> -j MASQUERADE
iptables -A FORWARD -i <LAN_IF> -o <WAN_IF> -j ACCEPT
iptables -A FORWARD -i <WAN_IF> -o <LAN_IF> -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
```

### 8. 疎通復旧確認

`client1`:

```bash
curl -s http://172.31.0.2 >/dev/null && echo client1_restore_ok
```

`client2`:

```bash
curl -s http://172.31.0.2 >/dev/null && echo client2_restore_ok
```

### 9. Step4 の成功判定

- 事前は疎通できる（`client1_precheck_ok`, `client2_precheck_ok`）
- ルール削除後は疎通できない
- ルール復旧後は疎通が戻る（`client1_restore_ok`, `client2_restore_ok`）
- カウンタや出力値の一致ではなく、`可 -> 不可 -> 可` の遷移で判定する

### 10. なぜそうなるか

- `conntrack` は状態管理を行いますが、変換や転送の実処理は `iptables` ルールが担います。
- ルールがない状態では、新規通信は変換/転送されず成立しません。
- ルールを復旧すると、新規通信が再び成立します。

### 11. 注意

- IF名は決め打ちせず、`ip addr` で確認してください。
- 出力やカウンタは環境で変わるため、値の一致ではなく構造で判定してください。
- ルール削除前に復旧コマンドを必ず控えてください。
