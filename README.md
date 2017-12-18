# coin-trade-history

仮想通貨の取引履歴をExcel形式で取得するツールです。
ZaifとbitFlyerに対応しています。
確定申告の助けになればと作成中です。

## 利用方法

[リリースページ](https://github.com/uphy/coin-trade-history/releases)からダウンロードしてください。  
続いてAPI Keyを取得して下さい。

取得場所
* bitFlyer: https://lightning.bitflyer.jp/developer
* Zaif: https://zaif.jp/api_keys

取得したら、API Key/Secretの内容を、config.ymlに書き込んでください。  
また、自分が取引した通貨をconfig.ymlのcurrenciesに列挙してください。  
(currenciesを削除すると全てを対象としますが、時間がかかります。)

例: ZaifでXEMとビットコインを取引した場合

```yml
- service: zaif
  ...
  currencies:
  - btc_jpy
  - xem_jpy
```

以上で利用準備完了です。以下のコマンドを実行して取引履歴をExcel形式で取得して下さい。

```
$ coin-trade-history download trade-history.xlsx
```

# 出力データについて

以下のようなデータを出力します。

![capture](https://raw.githubusercontent.com/uphy/coin-trade-history/images/sample.png)

それぞれの列について説明します。

| 列名 | 説明 |
|-----|----|
|Time|取引した日時です。|
|Service|サービス名です。bitFlyerもしくはZaifです。|
|Currency|仮想通貨の種類です。|
|Action|買った場合はBuy、売った場合はSellです。|
|Price|取引時の仮想通貨の単位価格(円)です。|
|Amount|仮想通貨の枚数です。|
|Fee|手数料です。|
|Profit|利益です。買った場合は`Price*Amount - Fee`、売った場合は`-Price*Amount - Fee`です。Zaifの場合、Bonusも考慮されます。|
|Total|これまでの利益の合算です。|
|Remarks|その他特記事項です。|

# 注意

- 開発途中です。正しく取得できない可能性があるので収支が合っているかよくご確認下さい。特に、取引履歴取得はページ処理をサボってるので、1万件しか取得しません。
1万件以上取引した方はご注意下さい。
- 本ツールで取得するのは取引履歴のみです。その他、サービスごとにあるログインボーナスやチャットボーナス等は考慮されません。
