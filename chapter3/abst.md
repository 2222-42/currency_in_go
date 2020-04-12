# Goにおける並行処理の構成要素

## 3.1 ゴルーチン(goroutine)

この節で扱うこと：ゴルーチンの概要、どのようにゴルーチンを起動するかについての説明。

この節での結論：ゴルーチンは安全に生成できるし、コストは非常に低いからアムダールの法則によってすごくスケールしやすいことになる。

全てのGoのプログラムには最低1つのゴルーチン、メインゴルーチンがある。

ゴルーチンは他のコードに対し並行に実行している関数のこと

ゴルーチンの実行の仕方はすごく簡単。

ゴルーチンとはいったい何で、どう動作するのかの解説をする。

ゴルーチンは実際どのように動いているか、OSスレッドかグリーンスレッド(仮想マシン (VM) によってスケジュールされるスレッド)か
-> ゴルーチンはOSスレッドではなく、またかならずしもグリーンスレッドではない。

ゴルーチンはコルーチン(coroutine)として知られる高水準の抽象化

コルーチンはプリエンティブでない並行処理のサブルーチン；つまり、割り込みをされることがない、かわりに、コルーチンには一時停止や再エントリーを許す複数のポイントがある

補足：

- プリエンプティブ: OSがCPUやシステム資源を管理し、CPU使用時間や優先度などによりタスクを実行状態や実行可能状態に切り替える方式。 (プロセスの切替えが頻繁に起こるので、コンテキスト切替えのオーバヘッドは大きくなります。)
- ノンプリエンプティブ: 実行プロセスの切替をプログラム自身に任せる方式で、プログラムが自発的にCPUを開放した時間でほかタスクを実行する。OSがCPUを管理しないので、1つのプログラムを実行中は、ほかのプログラムの実行は制限される。(特定のプロセッサがCPUを独占することは少なくなります)

ゴルーチンが特殊なコルーチンと考えられるがその独特さの所在は、ゴルーチンがGoのランタイムと密結合していることにある。一時停止や再エントリーのポイントを定義しておらず、ランタイムが自動でやってくれ、ゴルーチンがブロックしたら一時停止、ブロックが解放されたら再開として、ゴルーチンをプリエンプティブにしている。

並行性はコルーチン、そしてゴルーチンの性質ではない。コルーチンが暗黙的に並列であるということを示唆するわけではない。

Goがコルーチンをホストする機構は`M:Nスケジューラー`と呼ばれる実装、`M`個のグリーンスレッドを`N`個のOSスレッドに対応させるもの、になっている。ゴルーチンはグリーンスレッドにスケジュールされる。詳しくは6章で話す。

Goは`fork-joinモデル`と呼ばれる並行処理のモデルに従っている。mainから子の処理が分岐(fork)され、親と並行に実行され、並行処理の分岐が再び合流(join)する。合流する場所を合流ポイントという。

go文はGoがどう分岐を実行するかを表し、分岐されたスレッドを実行しているのはゴルーチン。ただし、ゴルーチンが生成されて、Goのランタイムにスケジュールされるが、実行する機会が得られるかは不明。

合流ポイントを作成するためには、メインゴルーチンと分岐したゴルーチンを同期しなければならない。

例として取り上げられているsyncパッケージのWaitGroupを使った実装。例でわかったこと：

1. `wg.Add(1) -> defer wg.Done -> wg.Wait`でホストしているゴルーチンが終了するまでメインゴルーチンをブロック
2. ゴルーチンはそれが作られたアドレス空間と同じ空間で実行する。
3. ゴルーチンが実行される前に、ループが終了してしまうと、変数がスコープ外のものになり、反復変数の意図に反しヒープに移されたメモリを見る。そのため、反復変数のコピーをクロージャーに渡して、ゴルーチンが実行されるようになるまでにループの各繰り返しから渡されたデータを操作できるようにする。

ゴルーチンの利点：ゴルーチンはお互いにアドレス空間を操作し、単純に関数をホストしているため、ゴルーチンを使うことは並行でないコードをかくことの自然な延長。Goのコンパイラはうまい具合に変数をメモリに割り当ててくれる。だから、開発者はメモリ管理ではなく問題空間に集中できる。

→　しかし、なんでも勝手にできるわけではない。複数のゴルーチンは同じアドレス空間に対して操作をするので、同期に関しては気にかけなければならない。

ゴルーチンの他の利点：くそ軽量で数キロバイト(私の環境だと8.836kbだったのだが？？？？)

補足：Goのガベージコレクターは何らかの理由でブロックした状態になっているゴルーチンを回収するようなことは何もしない　→　4.3 「ゴルーチンリークを避ける」の節で扱う。

コンテキストスイッチ：並行プロセスをホストしているものが別の並行プロセスに切り替えるために状態を保存しなければならないときに起こるもののこと

OS層では、レジスタの値、参照テーブル、メモリマップなどの保存の必要性があり、スレッドのコンテキストスイッチは非常にコストが高くなる。一方、ソフトウェア内だと比較的ずっとコストは小さい。ソフトウェアで定義したスケジューラーでは、ランタイムは何を、どのように、いつ永続化すべきかに関して、より多くの選択肢がある。

Goのベンチマークだとコンテキストスイッチは225ナノ秒しかかからないから、ゴルーチンを使う上では、それほど障壁にはならない。

## 3.2 syncパッケージ

### 3.2.1 WaitGroup

### 3.2.2 MutexとRWMutex

### 3.2.3 Cond

### 3.2.4 Once

### 3.2.5 Pool

## 3.3 チャネル(channel)

## 3.4 select文

## 3.5 GOMAXPROCSレバー

## 3.6 まとめ