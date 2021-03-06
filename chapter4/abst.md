# 4章 Goでの並行処理パターン

4章で扱うこと：前章までで説明したプリミティブを組み合わせ、システムをスケーラブルで保守可能に保つパターンにする方法について

空インターフェース型(`interface{}`)を使う理由：

1. 簡潔に例を書くため
2. ある状況においてパターンが何を実現しようとしているかがわかりやすくなるから(cf: 4.6)

Goのジェネレーターを使っていつでもこうしたコードを生成することができるし、必要な型を使ったパターンを生成できる。

## 4.1 拘束

複数の並行プロセス内で安全な操作をするための方法: 

- メモリを共有するための同期のプリミティブ(eg. `sync.Mutex`)
- 通信による同期(eg. チャネル)

安全な操作をする方法で、データの中身を認識する不可を軽減したり、クリティカルセクションを小さくする方法：

- イミュータブルなデータ(Goだったらメモリ内の値へのポインターの代わりに値のコピーを使うようなコード)
- 拘束によって保護されたデータ

同期が使えるのに拘束を使う理由：(同期に比べてコストが低いので)パフォーマンスの向上、開発者に対する可読性の向上(レキシカルな拘束は一般に理解しやすいものになる)。

拘束は情報をたった一つの並行プロセスからの見えられることを確実にしてくれる単純かつ強力な考え方。

拘束には2種類ある：

- アドホックな拘束：コミュニティやチーム、コードベースなどによって指定されている規則、また静的解析の実行によって達成した場合のこと(実現は難しい)
- レキシカルな拘束：レキシカルなスコープを使って適切なデータと並行処理のプリミティブだけを複数の並行プロセスが使えるように公開すること(間違ったアクセスを不可能にすること)(グローバルではなくチャネルの内部で宣言し関数の内部だけで閲覧できるようにする)

しかし、拘束をきちんと作るのはむずかしい場合があり、並行処理のプリミティブを使う必要が出るだろう。

## 4.2 for-selectループ

`for-select`ループは次のようなもの以外の何物でもない

```
for { // 無限ループまたは何かのイテレーションを回す
	select {
	// チャネルに対して何かを行う
	}
}
```

このパターン出現するシナリオ例

- チャネルから繰り返しの変数を送出する。
- 停止シグナルを待つ無限ループ(select文を抜けた後に割り込みできない処理を書くか、default節に書くか)

## 4.3 ゴルーチンリークを避ける

少ないとはいえゴルーチンもコストがかかり、(通常だとランタイムによってGCされるのだが、ゴルーチンの場合はことなり、)ゴルーチンはランタイムによってガベージコレクションされないため、プロセス内にほっておきたくない

ゴルーチンが終了に至るまでの流れ：

- ゴルーチンが処理を完了する場合
- 回復できないエラーにより処理を続けらえない場合
- 停止するように命令された場合

並行処理でゴルーチンはお互いの作業内容を知る必要はないが協調して動いている。
子のゴルーチンが処理をし続けるかべきかどうかは他の多くのゴルーチンを知らないといけない。
親のゴルーチンがそのコンテキスト全て知ることで、子のゴルーチンにキャンセル処理を行う。

ゴルーチンリークの例で、問題の軽減をしていく。

読み込みのケース: nil チャネルを渡しちゃて、プロセスが生きている限りメモリ内に残り続けちゃって、メモリ使用率をじわじわ高めちゃう。
　→　親から子にキャンセルのシグナルを送れるようにしよう。シグナルは `done`という名前の読み込み専用チャネルするのが慣習。

書き込みのケース: 書き込みをしようとするがブロックされ続ける状況でも同じように、リークが発生する。
　→　読み込みと同様にキャンセルのシグナルを送れるようにしよう。

明記したほうが良い規約：あるゴルーチンがゴルーチンの生成の責任をもっているならば、そのゴルーチンを停止できるようにする責任もある。

この周りの技術とルールについては、パイプラインとcontextパッケージの説でより詳しく扱う。停止させるやり方はゴルーチンの種類と目的によって異なるが、doneチャネルを渡すという基本に基づいている。

## 4.4 orチャネル

orチャネルパターン: (実行時にまとめるべきチャネルの数がわからない)1つ以上のdoneチャネルを1つdoneチャネルにまとめて、いずれかのチャネルが閉じたら、まとめたチャネルも閉じるようにする(メタチャネルっぽさがあるが、チャネルのチャネルではない?)

テキストでは再帰とゴルーチンを使って合成したdoneチャネルを作れる例を紹介している。
(ここでは、いずれかのチャネルから読み込めるかどうか・準備完了しているかを延々確認し、orDoneをcloseさせ、それをreturnさせることで実現している。)

コンパイル時に扱うdoneチャネルがいくつあるかわからないのであれば、そもそも他にdoneチャネルをまとめる方法はない。

orチャネルパターンは、システム内で複数のモジュールを組み合わせる際の継ぎ目として利用すると便利。

木構造を再帰的に形成し、いずれかで準備が完了していたら、キャンセル条件を満たしたことにし、複雑化をなくし、単純にこれらを組み合わせてコールスタックに伝えられる。

同様のことはcontextパッケージの節で紹介する。

(チャネルの数が定まらない時点でやばいので、そういうコードにならないようにしよう。)

### 補足

for-selectの中身のチャンネルが複数ある場合、チャネルがどんどんcloseされていくと、それにつれて読み込める確率がヘリ、読み込みにかかるスループットが高くなる。

このスループットを減らすためにもor-channelは活用することができる。

```
for{
  select {
    case c1 := <- channel1
    case c2 := <- channel2
    case c3 := <- channel3
    case c4 := <- channel4
  }
}
```

## 4.5 エラーハンドリング

Goは人気のある例外処理機構を採用しないことを決めた。エラーの伝播はアルゴリズムを考えるときと同じくらいの注意を払うべき。エラーはゴルーチンから返される値を構築するさいの第一級市民として取り扱得られるべき。

最も根本的な疑問：「誰がそのエラーを処理する責任を持つべきか」

並行処理プロセスの場合は、より複雑になる。なぜなら、独立して並行に処理が実行されているから。

一般的に並行プロセスは、エラーを、プログラムの状態を完全に把握していて何をすべきかをより多くの情報に基づいて決定できる別の個所(メインゴルーチン)へ、送るべき。

- エラーを入れる型を作る(取得されるであろう結果とエラーを対にする -> メインゴルーチンが何をすべきかの決定ができる)
- インスタンスを作り、そこに放り込んで、
- メインゴルーチンでゴルーチンから発生するエラーを賢く、そしてより大きなプログラムのコンテキストをすべて理解したうえで扱え。

ゴルーチンがエラーを生成するのであれば、それらは正常系の結果と強く結びつけて、正常系と同じ経路を使って渡されるべき。

## 4.6 パイプライン

パイプライン：システムの抽象化に使える道具、データストリームやバッチ処理を扱う必要があるときにとても強力

パイプラインはデータを受け取って、何らかの処理を行って、どこかに渡すという一連の作業にすぎない

パイプラインのステージ：パイプライン内での各操作のこと

ステージを独立させるので、懸念事項の切り分けが可能になる。
- 組み合わせ方をステージの修正とは独立して変更できる。
- 上流下流のステージと並行に行える
- 細かな処理をファンアウトさせたり流量制限をかけたりできる。

パイプラインのステージの性質：

- ステージは受け取るものと返すものが同じ型である(前段のステージの戻り値の型と後段のステージの入力の型が一致していれば問題ない)
- ステージは引き回せるように具体化されてなければならない(言語が開発者に概念を公開して直接扱えるようにするという意味；理由は関数シグネチャの型を持つ変数を定義できるから)

(関数シグネチャ (もしくは型シグネチャ、メソッドシグネチャ) は関数やメソッドの入力と出力を定義します。)

「パイプラインはモナドの部分集合」　←　嘘

手続き的なコードはデータのストリームを処理する際にパイプラインが提供してくれるような利点は提供してくれない

- バッチ処理：データの塊を一度に処理する
- ストリーム処理：要素を1つずつ受け取って、1つずつ渡すやり方

バッチ処理での各ステージでは、元データと同じ長さのスライスを新しく作成して計算結果を保存していることが意味するところは、
プログラム内のある瞬間に必要なメモリのフットプリント(プログラムが実行時に必要とするメインメモリの容量の大きさ)は
パイプラインの初めに渡したスライスのサイズの倍になるということ

ストリーム処理の各ステージでのメモリフットプリントはパイプラインの入力のサイズまで小さくなる。
しかし、パイプラインをforループ本体に入れて、rangeに重労働させている。これだと再利用しづらく、スケーラビリティに影響を与える。
また、ループの繰り返しごとにパイプラインをインスタンス化している。

(一般的な意味での「ジェネレータ」とはちょっと違う使い方なので、要注意)

### 4.6.1 パイプライン構築のためのベストプラクティス

チャネルはパイプラインのステージの性質を満たしているから、チャネルはパイプラインを構築する上でGoならではの姿に適合している。

パイプラインでよく使うものとして、個別の値の塊をチャネル上を流れるデータのストリームに変換してくれる類の関数は `ジェネレーター`と呼ばれるものがある。

for文に入れていたケースとの違い：

- チャネルを使っているので、パイプラインの終わりにrangeを使って値を取り出し、そしてここでの入力値と出力値は並行処理の文脈で安全、よって、各ステージを安全に実行できる
- 各ステージを並行に処理できるので、どのステージでも入力値だけを待てばよくなり、すぐに出力を送ることができる(4.7でこの事実は重要な分岐点となる)

プログラムが処理を終える前にdoneチャネルに対してcloseを読んだらどうなるか。
　→　パイプラインのステージがどのような状態でも、doneチャネルを閉じれば、(入力値のチャネルに対するrangeでの繰り返し処理でも、チャネルへの送信の処理がdoneチャネルとselect文を共有しているので)
パイプラインを伝播しており、強制的にパイプラインのステージを終了できる。

上記の通り、再帰的な処理が行われている。処理を外部から割り込み可能にしなければならない処理：

1. 一瞬で作ることができないデータ群の生成(Goの場合は十分に早いので検討しなくてよい)
2. 個々の値のチャネルへの送信(select文とdoneチャネルによって対応される。これでチャネルへの書き込み処理がブロックしている場合でもジェネレーターを割り込み可能にしている)

パイプラインの終端、最後のステージは帰納的に割り込み可能であることが保証されている。

(ベストプラクティス：
1. doneチャネルを用意する
2. generatorを先頭に置く
3. パイプラインのステージはdoneチャネルによって割り込めるようにしておく
4. パイプラインを作る
5. ちゃんとcloseするようにしておく。
)

### 4.6.2 便利なジェネレーターをいくつか

1. `repeat`
2. `take`
3. `take` + `repeatFn` + `rand`

空インターフェース型はGoでは一種のタブーとされているが、パイプラインのステージに関していえば、標準ライブラリとして`interface{}`型をつかうことは問題ない。(問題ないが、プロジェクトなどの閉じた環境では型を明示するのが望ましい)
理由：再利用可能なステージによって多くのパイプラインの利便性が得られる。扱っている型に関する情報は必要とせず、パラメータの引数の数の情報のみが必要

特定の型を扱う場合には、型アサーションを行うステージを用意できる。ただし、型アサーションを行うことの性能のオーバーヘッドは無視できる。

ジェネリックなステージと特定の型のステージとで比較すると、特定の型のステージの方が2倍速いが、パイプライン上で制約になるのはジェネレーターか計算量が多いステージのどちらか。
ディスクやネットワークからの読み込みの場合、ここで示したような性能のオーバーヘッドは大したものではなくなる。

(無視できるとはいうが、型シグネチャがあることによるコードの可読性におけるメリットや、静的解析によるコンパイル時のエラー検出などの恩恵が受けられないことには留意すべき)

ジェネリックな手法が気に食わないなら、ジェネレーターを生成できるGoジェネレーターを使えばいいんじゃない？

計算コストが高いステージと言えば、その影響をどのように減らせるか、その影響でパイプライン全体に流量制限がかかってしまわないか。この影響を低減させる方法がファンアウト、ファンイン。

## 4.7 ファンアウト、ファンイン

計算量が多いステージがあると、上流のステージはブロックされてしまい、パイプライン全体の実行に時間がかかる。

パイプラインのステージを複数回使ったり、上流のステージから並列に値を引っ張ってきたりできる。

- ファンアウト：パイプラインからの入力を扱うために複数のゴルーチンを起動するプロセス
- ファンイン：複数の結果を1つのチャネルに結合するプロセス(マルチプレキシングから逆多重化を考え無くしたケース)

ファンアウトの利用を検討する場合：　

- そのステージがより前の計算結果に依存していない
- 実行時間が長時間に及ぶ

ファンアウトの方法、複数起動させてチャネルの配列を作って、それぞれで作業させるだけ。

ファンインの方法：
1. 消費者の読み込み先となるマルチプレキシングしたチャネルを作成
2. その後入力値となるチャネル各々に対しゴルーチンを起動し
3. 入力値となるゴルーチンがすべて閉じられたら、多重化したチャネルを閉じる、ためのゴルーチンを起動する

補足：ファンインとファンアウトのナイーブな実装は結果の順序が重要でない場合のみにのみうまく動作。順序を維持する方法についてはのちに見る。

通常の結果：
```
Primes:
        48498081
        27131847
        39984059
        11902081
        24941318
        40954425
        36122540
        8240456
        46203300
        6410694
Search took: 129.9993ms

```

ファンアウトとファンインを使った場合の結果：
```
Primes:
Spinning up 12 prime finders.
        48498081
        27131847
        11902081
        39984059
        24941318
        8240456
        36122540
        6410694
        46203300
        10128162
Search took: 2.9997ms
```


## 4.8 or-doneチャネル

システムの完全に異なる部分から受け取ったチャネルを扱う場合、
パイプラインと違い、doneチャネル経由でキャンセルされた場合に受け取ったチャネルがどのようにふるまうか判断出来ない。

ゴルーチンがキャンセルされたという事実が、読み込み先のチャネルがキャンセルされたという意味になるかもしれない。
だから、loop文を回しそうになるが、select文を連続して書きましょう。

```
orDone := func(done, c <-chan interface{}) <-chan interface{}{
	valCh := make(chan interface{})
	go func() {
		defer close(valCh)
		for {
			select{
			case <-done:
				return
			case v, ok := <-c:
				if !ok {
					return
				}
				select{
				case valCh <-v:
				case <-done:
				}
				
			}
		}
	}()
	return valCh
}

for val := range orDone(done, myChan) {
// do something
}
```

## 4.9 teeチャネル

チャネルからのストリームを二つに分けて、同じ値を2つの異なる場所で使わせたい場合

元のチャネルからの繰り返しの読み込みは、書き込み先の2つのチャネルへの書き込みが終わらない限り進まないように。また、スループットはteeコマンド以外の何かの影響が大きいので、このことは問題ない。

## 4.10 bridgeチャネル

チャネルのチャネル( `<-chan <-chan interface{}` )でチャネルのシーケンスから値を取りたい場合がるだろう。

この場合は、チャネルのスライスを1つのチャネルにまとめるのではなく、チャネルのシーケンスであり、複数のリソースからであっても書き込み順を提示できる。

チャネルのチャネルを扱うのは面倒だから、チャネルを崩して単一のチャンネルにする、ブリッジングと呼ばれる関数を定義しよう。

## 4.11 キュー

あるステージでの処理が終わった時に、メモリに一時的にその結果を保存して、あとで他のステージがそれを取得できるようにし、先のステージが値を参照し続けないで済むようにする。

バッファ付きチャネルは一種のキュー。

キューの役立つ場面：

- ステージがブロック状態になっている時間が短くなること。これによって、ステージが処理を続けられる。

キューを導入する際のよくある誤り:

- キューはプログラムの合計の実行時間はほとんど改善させません。プログラムに違った振る舞いをさせるだけ。

キューの真の実用性：

- あるステージの実行時間がほかのステージの実行時間に影響を与えないようにステージを分離すること

キューはどこに置くべきか、バッファのサイズはどう設定すべきか　→　パイプラインの性質に依存した答えになる

(キューはどこに置くべきか)キューがシステム全体の性質を向上させうる状況は以下の二つ:

1. ステージ内でのバッチによるリクエストが時間を節約する場合(送信先より早いものからの入力をバッファする場合(e.g. Goの`bufio`パッケージ))　→　バッチ処理によって効率的になるステージの中
2. ステージにおける遅延がシステムにフィードバックループを発生させる場合　→　パイプラインの入り口

Goの`bufio`パッケージを使った場合、バッファありの方が早いのは、 `bufio.Writer`の中で書き込むに十分な量が蓄積されるまでバッファに待ち合わせて、
その後にまとめて書き込んでいるため。(チャンキングと呼ばれる。オーバーヘッドを必要とする処理を行うならこれはシステムの性能の向上につながる)
これでシステムコールの呼び出し回数が減って、速くなる。

1についてはチャンキング以外にも、あと夜宮順序付けをサポートすることで最適化できる場合にもキューが役立つ。

2については、見分けるのは難しいが、1よりも重要。重要な理由は、上流のシステムを全体的に崩壊させる可能性があるから。

2はネガティブフィードバックループ、下方スパイラル、デススパイラルとも呼ばれる。パイプラインと上流のシステムの間に再帰的な関係がある場合に発生する。

パイプラインの入り口にキューを導入することで(呼び出し元から見ればリクエストは処理されているように見える)
リクエストに対する時間差を発生させる代わりに
フィードバックループを崩す。(呼び出し元がタイムアウトした場合は、フィードバックループを形成してパイプラインの効率を下げる事態を避けるために、キューから取り出す際に呼び出し元が用意できているか(死んだリクエストではないか)を確実に確認する何かしらの対応をしている必要がある。)

パイプラインのスループットについて予測する方法がある。あるリクエスト数を処理するために必要なキューの大きさをどう決定するかもここからわかる。

リトルの法則 : `L = λW`

- L: システムの平均ユニット数
- λ: ユニットの平均の到達率
- W: ユニットのシステム内での平均滞在時間

ここからわかることの例：
1. ユニットのシステム内で平均滞在時間を減らしたい
2.  -> システム内の平均ユニット数を減らすしかない。(L/n = λ * (W/n))
3.  -> 流出速度を上げるしかあに

ステージにキューを追加することが意味すること：
 -> これはユニットの到達率を上げる(nL = nλ * W)、もしくは滞在時間を増やすことになる。(nL = λ * nW)
 
リトルの法則によって、キューはシステムの実行時間を減らす助けにならないことがわかる。

パイプラインのステージ全てを通して分散というのは、`L = λΣ_i W_i`ということ。

安定したシステム(パイプラインの流入の速度と流出の速度が等しい場合)にのみ適用可能な式であり、失敗を扱う場合は成立しない。

リクエストの再生成が困難だったり、二度とできないような場合には、キュー内のリクエスト全て失いことを防ぐためにどうするか：

- キューの大きさをゼロにする
- 永続キューに移行してもよい

キューはシステム内で便利だが、複雑なので、筆者は実装する際には最適化の最後の手段として提案する。

## 4.12 contextパッケージ

### 用語

- コールスタック (Call Stack)は、プログラムで実行中のサブルーチンに関する情報を格納するスタックである
  - プログラムの現在位置 (ブレークポイント、ステップ実行のあと、プログラムが異常終了してコアファイルが作成された、いずれかの時点で実行されていたルーチン)
   はメモリー上位に存在しますが、main() のような呼び出し側ルーチンはメモリー下位に位置します。
   (cf: https://docs.oracle.com/cd/E19205-01/820-1199/blafv/index.html)
- コールグラフ(マルチグラフとも。コンピュータプログラムのサブルーチン同士の呼び出し関係を表現した有向グラフである。
具体的には、各ノードが手続きを表現し、各エッジ(f,g)は手続きfが手続きgを呼び出すことを示す。
従って、循環したグラフは再帰的な関数呼び出しを示す　from: WIKIPEDIA)
- プリエンプション（英: preemption）は、マルチタスクのコンピュータシステムが実行中のタスクを一時的に中断する動作であり、
基本的にそのタスク自体の協力は不要で、後でそのタスクを再実行するという意味も含む。(from WIKIPEDIA)

### 疑問

テキストで上位や下位、上、下、下流などの言葉が使われているが、この区別はちゃんとされているか？

もしされているとしたらどういう意図で使われているか、それを明らかにしましょう。

### contextパッケージについて

doneチャネルはプログラム全体を流れ、ブロックしている並行処理をすべてキャンセルする。

単純なキャンセルの通知に付随して以下のような追加の情報を伝達できたらいいな:

- キャンセルが発生した理由や
- 関数の処理を終わらせるべきデッドライン(`Deadline`)

Context型について:

- doneチャネルのようにシステム内を流れる
- contextパッケージを使う場合には、並行処理の呼び出し元の最上位より下流の各関数は `Context`を第一引数として受け取る(doneチャネルの慣習と一緒)

Context型の定義：

- Deadline() : ゴルーチンが一定の時刻以降にキャンセルされるかを返す
- Done() : 関数がランタイムにより割り込みされたときに閉じるチャネルを返すメソッド
- Err() : ゴルーチンがキャンセルされたら非nilな値を返す
- Value(key interface{})

Valueメソッドの目的について：リクエストをさばくプログラムでは、ランタイムの割り込みに関する情報に加えてリクエストに応じた情報が渡される必要があるから

contextパッケージの主要な目的:
1. コールグラフの各枝をキャンセルするAPIを提供する
2. コールグラフを通じてリクエストに関するデータを渡す

### contextパッケージの目的1：キャンセル編

以下の関数内でのキャンセルの3つの側面、いずれの場合でもcontextパッケージは役立つ:

- ゴルーチンの親がキャンセルをしたい場合
- ゴルーチンが子をキャンセルしたい場合
- ゴルーチン内のブロックしている処理がキャンセルされるよう中断できる必要がある場合

コールスタックの上位の関数が下位の関数によってコンテキストをキャンセルされることはない(上位と下位はどっちがどっち？)。
理由は、Contextインターフェースには以下の特徴があるから:

1. 内部構造の状態を変更できるものはない
2. Contextを受け取ってそれをキャンセルさせられる関数もない。

これによって、doneチャネルを提供するDoneメソッドと組み合わせることで、Context型がその祖先からキャンセルを安全に管理できるようになる。

Contextはイミュータブル　→　コールスタックないでいまいる関数より下(下位？子？)の関数に対するキャンセルによる振る舞いの影響を与えられるか

→　ある関数がコールグラフ内でそれ以降の関数をキャンセルする必要がある場合、以下のcontextパッケージの関数で、
Contextを第一引数にとり、新たなContextのインスタンスを生成し、それを子の関数に渡せばよい。
キャンセルする必要がない場合は、元のContextをそのまま子に渡せばよい。

- WithCancel(parent Context) (ctx Context, cancel CancelFunc): cancel関数が呼ばれたときにそのdoneチャネルを閉じる新しいContextを返す
- WithDeadline(parent Context, deadline time.Time) (Context, CancelFunc): マシンの時計が与えられた時刻を経過したらそのdoneチャネルを閉じる新しいContextを返す
- WithTimeout(parent Context, timeout time.Time) (Context, CancelFunc): 与えられた時間だけ経過したらそのdoneチャネルを閉じる新しいContextを返す

→　各レイヤに付随する要求に関するContextを親への影響を与えることなく作成可能

context.Contextのインスタンスは外から見たら同等に見えるかもしれないが、内部的にはスタックフレームごとに変化する(参照は保管しておかない)。
　→　毎回関数の引数にContextを入れることが重要。

非同期なコールグラフの場合、おそらくContextは渡されない。

Contextのからインスタンスを作る関数2つ:
1. func Background() Context : 空のコンテキストを返す
2. func TODO() Context : 本番環境で使うことは想定していない、どのContextを使っていいかわからないときや、上流のコードの実装がまだ終わっていないときのプレースホルダー

並行で処理される2つの枝があるコードの例：
― doneチャネルをmainのどの場所で閉じても枝はりょうほうともキャンセルされる。
- doneチャネルの場合、各スタックフレームにおいて、関数はそれ以後のコールスタック全体に影響を与える。
  - doneの場合は、親がすぐにキャンセルされるなら、新たな関数を呼び出させたくないだろう。
  - doneの場合は制限時間やエラーに関する情報は得られない
- contextを使えばいいじゃない

親への影響を与えることができずに独自の`context.Context`を作っている。
この構成可能性によってコールグラフ全体を通じて懸念事項をまぜこぜにすることなく(独自のルールはその関数の中に入れられる;これはdoneの場合にはないメリット)、
大きなシステムを記述できる

実行に一定時間以上かかることがわかっている、後続のコールグラフがどれくらいかかるかをある程度わかっているのであれば(この条件の成立は難しいが)、
context.ContextのDeadlineメソッドを使って、以下のことが可能になる。
- デッドラインがあるのかどうかの確認
- あるのならそれに対応するかどうかの確認をすることが可能

例：デッドラインを利用したケース

これによって、早く失敗することができる。これは、呼び出すコストが高いプログラムにおいては、時間の短縮が可能になる。

### contextパッケージの目的2：データバッグ編

Contextのためにあるリクエストの範囲でのデータを保管し受け取るためのデータバッグでもある。

リクエストを出すプロセスを起動して、それ以下の関数はそのリクエストに対する情報が必要になるので、それをContext内で保管して、それを受け取る例

使い方とリマーク：
- `context.WithValue(ctx, :key, :value)`で設定できる
- 満たさなければいけないこと:
  - 使用するkeyはGoでの比較可能性を満たさなければならないこと
  - 返された値は複数のゴルーチンからアクセスされても安全でなければならないこと
― Contextのキーと値はinterface{}型として定義されているので、値を受け取るときはGoの型安全性は損なわれる。→　対応策がある。

以下の通り、Contextから型安全に値を受け取る方法があり、これにより値を受け取る側が情報の保管について使われたキーを知ることができる。

パッケージ内で独自のキーの型を定義することが推奨される。
→　キーの元の値がが同じだけれど、マップ内では型情報によって区別される。

キー用に定義した型はエクスポートされないので(privateにするので)、ほかのパッケージがこのキーと名前とが衝突することはない。
→　データの取得に関する関数をエクスポートする必要がある。

上記方法の問題点：
- 循環参照が起きる可能性がある

→　循環参照を避けるため、複数の場所からインポートされるデータ型を中心としたパッケージを作るようなアーキテクチャを強いられることになる。

### contextパッケージを使うための経験則

contextパッケージは型安全性の欠如を伴うから、好かれない場合もある。
　→　筆者は型安全性やバグよりも、何をContextのインスタンスに補完すべきかという性質の方がより大きな問題とみなす。

context packageにあるコメント:
「コンテキスト値はプロセスやAPIの境界を通過するリクエストスコープでのデータに絞って使いましょう。
関数にオプションのパラメーターを渡すためにつかうべきではありません。」

「リクエストスコープでのデータ」とは何か、多くのことを表現しているから、定義するならチームの経験則により江良らえるものをコードレビューの際に評価するというもの。

データをContextの保管するかどうかの筆者の経験則5つ(チームによっては当然異なる):

1. データはプロセスやAPIの境界を通過すべき:　プロセス内のメモリでデータを生成し、かつ、そのデータをAPIの境界を越えて渡さないようなものなら、不適格。
2. データは不変であるべき: 変化するなら、それはリクエスト由来の者ではない
3. データは単純な型に向かっていくべき: 通過した先にいる側がこのデータをずっと簡単に取得できるはずだから
4. データはデータであるべきでメソッド付きの型であるべきではない: 操作はロジックに属し、データを消費するものが操作だから。
5. データは修飾の操作を助けるべきものであって、それを駆動するものではない: オプションのパラメーターが扱うべき領域を犯しているから

他に考えるべきこと：データが使われるまでに何層をまたぐ必要があるか
→　これによって、以下のどちらかよいかはチームの意思決定による：

- 言葉数の多い自己説明的な関数シグネチャに寄せてデータをパラメーターとして渡すか
- Context内にデータを置く場所を確保して、見えない依存関係を作るか

リクエストID：リクエストごとのIDのこと
(APIの境界を通過する、不変、単純でありそのまんま、何もしないメソッドなし、ログとかに吐き出す修飾用)

ユーザーID
(APIの境界を通過する、不変、単純でありそんまんま、何もしないメソッドなし、ユーザーIDによって操作内容が変わる)

URL：
(APIの境界を通過する、不変、関数などに置き換えられる、メソッドあり、操作内容が変わる)

APIサーバーの接続：
(通過しない、可変、インフラ層で複雑なまま、それで駆動内容が変わる)

認可トークン：
(APIの境界を通過する、不変、単純でありそんまんま、何もしないメソッドなし、AuthTokenによって操作内容が変わる)

リクエストトークン(リサービスプロバイダーがアプリケーションに対して一時的に発行するもの。ユーザーがアプリケーションを承認するプロセスでのみ利用する。リサービスプロバイダーがアプリケーションに対して一時的に発行するもの。ユーザーがアプリケーションを承認するプロセスでのみ利用する。)：
(APIの境界を通過する、不変、単純でありそんまんま、メソッド付きの型ではない？？？、リクエストトークンによって操作内容が変わる)

教訓：
これらはチームによるので、使い方に関してはある種の意見を持とう

## 4.13 まとめ

Goの並行処理のプリミティブを組み合わせてパターンを作り、保守しやすい並行処理のコードを書きやすくしている。

この後はこれらのパターンをどのように別のパターンに組み込んで、大きなシステムの実装に役立てるか　→　chapter 5
