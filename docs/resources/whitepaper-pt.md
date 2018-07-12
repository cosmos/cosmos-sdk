# Cosmos

Uma Rede de Distribuição de Ledgers

Jae Kwon jae@tendermint.com<br/>
Ethan Buchman ethan@tendermint.com

Para discussões, [entre no nosso Matrix](https://riot.im/app/#/room/#cosmos:matrix.org)!

_NOTA: Se você pode ler isso no GitHub, então ainda estamos desenvolvendo este documento ativamente. Por favor, cheque regularmente as atualizações!_

\[[toc]]

O sucesso combinado do ecossistema de código aberto, compartilhamento
de arquivos descentralizado e criptomoedas públicas tem inspirado um conhecimento sobre
protocolos descentralizados na Internet que podem ser utilizados para melhorar radicalmente
a infraestrutura. Vimos aplicações de blockchain especializadas como Bitcoin
[\[1\]][1] (uma criptomoeda), Zerocash [\[2\]][2] (uma criptomoeda para privacidade
), and generalized smart contract platforms such as Ethereum [\[3\]][3],
com inúmeras aplicações distribuídas para a Etherium Virtual Machine (EVM), como Augur (uma previsão
de mercado) e TheDAO [\[4\]][4] (um clube de investimento).

Contudo, até à data, estas blockchains sofreram uma série de inconvenientes,
incluindo sua ineficiência energética, desempenho fraco ou limitado e
mecanismos de governança imaturos. Propostas de escala
de processamento de transações da Bitcoin, como Testemunhas Separadas [\[5\]][5] e
BitcoinNG [\[6\]][6], soluções de escalonamento vertical que permanecem
limitadas pela capacidade de uma única máquina física, a fim de
proporcionar uma auditabilidade completa. A Rede Lightning [\[7\]][7] pode ajudar
o Bitcoin no quesito volume de transações, deixando algumas transações completamente
fora da carteira, e é bem adequado para micropagamentos e preservando a privacisadade por pagamentos
Rails, mas pode não ser adequado para necessidades de escala mais abrangente.

Uma solução ideal é a de permitir blockchains paralelos múltiplos para
interoperação, mantendo suas propriedades de segurança. Isto provou
ser difícil, se não impossível, com prova de trabalho. A mineração combinada, por exemplo,
permite que o trabalho feito para proteger uma blockchain mãe seja reutilizado em uma blockchain nova,
mas as transações ainda devem ser validadas, em ordem, por cada nó, e uma
blockchain Merge-mined é vulnerável a ataques se a maioria do poder de
hashing sobre a mãe não é ativamente merge-mined da nova. Uma revisão acadêmica
do [arquiteturas de redes alternativas blockchain
](http://vukolic.com/iNetSec_2015.pdf) é fornecida para
contextualizar, e fornecemos resumos de outras propostas e suas desvantagens em
[Trabalho relatado](#trabalho-relatado).

Nesse relato nós apresentamos a Cosmos, uma novela da arquitetura de rede blockchain que aborda todos
esses problemas. Cosmos é uma rede de muitos blockchains independentes, chamados
Zonas. As zonas são alimentadas pelo Tendermint Coreork [\[8\]][8], que fornece uma
alta performace, consistência, segurança
[PBFT-como](https://blog.cosmos.network/tendermint-vs-pbft-12e9f294c9ab?gi=20a63f2a00ee) um mecanismo de consenso
rigoroso, onde [fork-responsável](#fork-responsável) tem-se garantias de deter
comportamentos maliciosos. O algoritmo de consenso BFT do Tendermint Core é
bem adaptado para integrar blockchains públicas de prova de estaca.

A primeira zona na Cosmos é chamada de Cosmos Hub. A Cosmos Hub é uma criptomoeda
multi-asset de prova de estaca com um simples mecanismo de governança
o qual permite a rede se adaptar e atualizar. Além disso, a Cosmos Hub pode ser
extendida por conexão com outras zonas.

O hub e as zonas da rede Cosmos comunicam-se uma com a outra através de um
protocolo de comunicação Inter-blockchain (IBC), um tipo de UDP ou TCP virtual para
blockchains. Os tokens podem ser transferidos de uma zona para outra com segurança e
rapidez sem necessidade de liquidez cambial entre as zonas. Em vez disso, todas
as transferências de tokens inter-zonas passam pelo Hub Cosmos, que mantêm
a quantidade total de tokens detidas por cada zona. O hub isola cada zona da
falha das outras zonas. Porque qualquer um pode conectar uma nova zona no Hub Cosmos,
o que permite futuras compatibilidades com novas blockchains inovadoras.

## Tendermint

Nesta seção, descrevemos o protocolo de consenso da Tendermint e a interface
usada para construir aplicações através dele. Para mais detalhes, consulte o [apêndice](#apêndice).

### Validadores

No algorítimo de tolerância e falhas clássicas Bizantinas (BFT), cada node tem o mesmo
peso. Na Tendermint, nodes tem uma quantidade positiva de _poder de voto_, e
esses nodes que tem poder de voto positivo são chamados de _validadores_. Validadores
participam de um protocolo de consenso por transmissão de assinaturas criptográficas,
ou _votos_, para concordar com o próximo bloco.

Os poderes de voto dos validadores são determinados na gênese, ou são alterados
de acordo com a blockchain, dependendo da aplicação. Por exemplo,
em uma aplicação de prova de participação, como o Hub da Cosmos, o poder de voto pode ser
determinado pela quantidade de tokens usados como garantia.

_NOTA: Frações como ⅔ e ⅓ referem-se a frações do total de votos,
nunca o número total de validadores, a menos que todos os validadores tenham
peso._
_NOTE: +⅔ significa "mais do que ⅔", enquanto ⅓+ significa "⅓ ou mais"._

### Consenso

Tendermint é um protocolo de consenso BFT parcialmente sincronizado e derivado do
algoritmo de consenso DLS [\[20\]][20]. Tendermint é notável por sua simplicidade,
desempenho, e [fork-responsável](#fork-responsável). O protocolo
requer um grupo determinado de validadores, onde cada validador é identificado por
sua chave pública. Validadores chegarão a um consenso em um bloco por vez,
onde um bloco é uma lista de transações. A votação para o consenso sobre um bloco
acontece por rodada. Cada rodada tem uma líder-de-rodada, ou proponente, que propõe um bloco. Os
validadores, em seguida, votam, por etapas, sobre a aceitação do bloco proposto
ou passam para a próxima rodada. O proponente de uma rodada é escolhido
de acordo com uma lista ordenada de validadores, proporcionalmente à seu
poder de voto.

Os detalhes completos do protocolo estão descritos
[aqui](https://github.com/tendermint/tendermint/wiki/Byzantine-Consensus-Algorithm).

A segurança da Tendermint é baseada na tolerância e falhas clássicas Bizantinas ótimas
através de super-maioria (+⅔) e um mecanismo de bloqueio. Juntas, elas garantem
isso:

- ⅓+ o poder de voto deve ser bizantino devido a violações de segurança, onde mais
    que dois valores são comprometidos.
- se algum conjunto de validadores tiver sucesso em violar a segurança, ou mesmo tentarem
  para isso, eles podem ser identificados pelo protocolo. Isso inclui tanto o voto
  para blocos conflitantes quanto a transmissão de votos injustificados.

Apesar de suas fortes garantias, a Tendermint oferece um desempenho excepcional. Dentro
de Benchmarks de 64 nós distribuídos em 7 datacenters em 5 continentes, em
nuvens de commodities, o consenso da Tendermint pode processar milhares de
transações por segundo, com tempo de resposta entre um a dois
segundos. Notavelmente, o desempenho muito além de mil transações por segundo
é mantido mesmo em condições adversas, com validadores falhando ou
combinando votos maliciosamente. Veja a figura abaixo para mais detalhes.

![Figura do desempenho da Tendermint](https://raw.githubusercontent.com/gnuclear/atom-whitepaper/master/images/tendermint_throughput_blocksize.png)

### Clientes Light

O principal benefício do algoritmo de consenso da Tendermint é um cliente leve e simplificado
de segurança, tornando-o um candidato ideal para o uso de dispositivos móveis e casos de uso na
internet. Enquanto um cliente leve do Bitcoin deve sincronizar blockchains e encontrar
o que tem mais prova de trabalho, os clientes light da Tendermint precisa apenas
das alterações feitas pelo conjunto dos validadores, em seguida, verifica-se o +⅔ PreCommits
no último bloco para determinar o estado atual.

Provas claras e sucintas do cliente também permite [comunicação-inter-
blockchain](#comunicação-inter-blockchain-ibc).

### Previnindo ataques

A Tendermint dispõe de medidas de proteção para evitar
ataques, como [gastos duplos em longa-distância-sem-estaca double
spends](#previnindo-ataques-de-longa-distância) e
[censura](#superando-forks-e-censurando-ataques). Esses são discutidos
completamente no [apêndice](#apêndice).

### TMSP

O algoritmo de consenso Tendermint é implementado através de um programa chamado Tendermint
Core. O Tendermint Core é um "mecanismo de consenso" independente de aplicações que
transformam qualquer aplicação blackbox em uma réplica distribuída na
Blockchain. Tendermint Core conecta-se ao blockchain
através de aplicações do Tendermint Socket Protocol (TMSP) [\[17\]][17]. Assim, o TMSP
permite que as aplicações da blockchain sejam programadas em qualquer idioma, não apenas
a linguagem de programação que o mecanismo de consenso é escrito, além disso,
o TMSP torna possível a troca fácil da camada de consenso de qualquer
tipo de blockchain.

Nós fizemos uma analogia com a bem conhecida criptogradia do Bitcoin. Bitcoin é uma
blockchain de criptomoedas onde cada nó mantém uma Unspent totalmente auditada
e banco de dados de saída de transação (UTXO). Se alguém quisesse criar um Bitcoin-like
TMS, a Tendermint Core seria responsável por

- Compartilhar blocos e transações entre os nós
- Estabelecer uma ordem de transações canônica/imutável (a blockchain)

Entretanto, o aplicativo TMSP seria responsável por

- Manter o banco de dados UTXO
- Validar a criptografia das assinaturas das transações
- Previnir transações vindas de gastos de fundos não exisentes
- Permitir aos clientes a consulta do banco de dados UTXO

Tendermint é capaz de decompor o design da blockchain, oferecendo um simples
API entre o processo da aplicação e o processo do consenso.

## Visão Geral da Cosmos

Cosmos é uma rede de blockchains paralelos e independentes que são alimentadas pelo
clássico algorítimo de consenso BFT como a Tendermint
[1](https://github.com/tendermint/tendermint).

A primeira blockchain dessa rede será a Cosmos Hub. A Cosmos Hub
conecta as outras blockchains (ou _zonas_) através do protocolo de comunicação-inter-
blockchain. A Cosmos Hub rastreia vários tipos de tokens e mantém
registo do número total de tokens em cada zona ligada. Os tokens podem ser
transferidos de uma zona para outra de forma segura e rápida, sem necessidade de
uma troca líquida entre zonas, porque todas as transferências de moedas ocorre
através da Cosmos Hub.

Essa arquitetura resolve muitos dos problemas enfrentados atualmente pelas blockchains,
tais como interoperabilidade de aplicativos, escalabilidade e capacidade de atualização contínua.
Por exemplo, as zonas baseadas do Bitcoin, Go-Ethereum, CryptoNote, ZCash, ou qualquer
sistema blockchain pode ser ligado ao Cosmos Hub. Essas zonas permite a Cosmos
o escalonamento infinito para atender a demanda global de transações. As Zonas também são um grande
apoio para a exchange distribuída, que também serão apoiadas.

Cosmos não é apenas uma única ledger distribuídos, o Cosmos Hub não é um
jardim cercado ou o centro do universo. Estamos elaborando um protocolo para
uma rede aberta de legers distribuídos que pode servir como um novo
futuros para sistemas financeiros, baseados em princípios de criptografia, economia
teoria de consenso, transparência e responsabilidade.

### Tendermint-BFT

O Cosmos Hub é a primeira blockchain pública na rede Cosmos, alimentada pelo
algoritimo de consenso BFT Tendermint. A Tendermint é um projeto de fonte aberta que
nasceu em 2014 para abordar a velocidade, a escalabilidade e as questões
do algoritimo de consenso da prova-de-trabalho do Bitcoin. Usando e melhorando
algoritmos BFT comprovados e desenvolvidos no MIT em 1988 [\[20\]][20], o time Tendermint foi o primeiro a
que demonstrou conceitualmente uma prova de estaca das criptomoedas que aborda o
problema de "sem-estaca" sofrido pelas criptomoedas da primeira geração
tais como NXT e BitShares.

Hoje, praticamente todas carteiras móveis de Bitcoin usam servidores confiáveis, que fornece
a elas transações com verificação. Isso porque a prova-de-trabalho exige
muitas confirmações antes que uma transação possa ser considerada
irreversivel e completa. Os ataques de gasto-duplo já foram demonstrados em
serviços como a CoinBase.

Ao contrário de outros sistemas de consenso blockchain, a Tendermint oferece
comprovação segura de pagamento para o cliente móvel. Uma vez que a Mint é
projetada para nunca passar por um fork, carteiras móveis podem receber confirmações de transações
instantâneas, o que torna os pagamentos confiáveis e práticos através de
smartphones. Isto tem implicações significativas para as aplicações da Internet.

Validadores na Cosmos tem uma função similar aos mineiros do Bitcoin, mas usam
assinaturas criptografadas para votar. Validadores são máquinas seguras e dedicadas
que são responsáveis por validar os blocos. Os não validadores podem delegar através de seus tokens estacados
(chamados "atoms") a qualquer validador para ganhar uma parcela das taxas da blockchain
e recompensas de atoms, mas eles correm o risco de serem punidos (cortados) se o
o validador de delegados for invadido ou violar o protocolo. A segurança comprovada
garantida pelo consenso BFT da Tendermint, e o depósito de garantia das
partes interessadas - validadores e delegados - fornecem dados prováveis,
segurança para os nós e clientes light.

### Governança

Ledgers de distribuição pública devem ser constituídos de um sistema de governança.
O Bitcoin confia na Fundação Bitcoin e na mineração para
coordenar upgrades, mas este é um processo lento. Ethereum foi dividido em ETH e
ETC depois de hard-fork para se recuperar do hack TheDAO, em grande parte porque não havia
contrato sócial prévio, nem um mecanismo para tomar tais decisões.

Os validadores e os delegados do Cosmos Hub podem votar propostas que
alteraram automaticamente os parâmetros predefinidos do sistema (tal como o gás limite do
bloco), coordenar upgrades, bem como votar em emendas para a
constituição que governa as políticas do Cosmos Hub. A Constituição
permite a coesão entre as partes interessadas em questões como o roubo
e bugs (como o incidente TheDAO), permitindo uma resolução mais rápida e mais limpa.

Cada zona pode ter sua própria constituição e mecanismo de governança.
Por exemplo, o Cosmos Hub pode ter uma constituição que reforça a imutabilidade
no Hub (sem roll-backs, a não ser por bugs em implementações dos nós do Cosmos Hub),
enquanto cada zona pode ter sua própria política sobre os roll-backs.

Ao disponibilizar a interoperabilidade em diferentes políticas das zonas, a rede Cosmos
dá aos usuários total liberdade e potenciais permissões para
experimentos.

## O Hub e as Zonas

Aqui nós descrevemos o modelo do roteiro de descentralização e ecalabilidade. Cosmos é uma
rede de muitas blockchains alimentadas pela Tendermint. Enquanto existirem propostas visando
criar um"blockchain solitário" com ordens de transações cheias, a Cosmos
permite que muitas blockchains rodem junto de outra enquanto mantêm a
interoperabilidade.

Basicamente, o Cosmos Hub gerencia várias blockchains independentes chamadas "zonas"
(as vezes chamadas de "shards", em referência a técnica de escalonamento de
bando de dados conhecida como "sharding"). Uma constante transmissão de blocos recentes das
zonas atuando no Hub permite ao Hub manter o estado de cada zona atualizado.
Sendo assim, cada zona mantêm ativa o estado do Hub (mas as zonas não se mantêm ativas
com qualquer outro exceto o Hub). Pacotes de informação são
então comunicados de uma zona para outra atráves de Merkle-proofs como evidências,
essas informações são enviadas e recebidas. Esse mecanismo é chamado de
comunicação inter-blockchain, ou IBC para encurtar.

![Figura de reconhecimento
do hub e das zonas](https://raw.githubusercontent.com/gnuclear/atom-whitepaper/master/images/hub_and_zones.png)

Qualquer uma das zonas podem ser hubs para formar gráficos acíclicos, mas
mas para deixar claro, nós vamos apenas descrever uma simples configuração para
um único hub, e várias zonas que não são hubs.

### O Hub

O Cosmos Hub é uma blockchain que hospeda um ledger de distribuíção de multi-asset,
onde os tokens podem ser mantidos por usuários individuais ou pelas próprias zonas. Esses
tokens podem ser movidos de uma zona para outra em um pacote IBC especial chamado
"coin packet". O hub é responsavel por preservar a manutenção global de
toda a quantia de cada token nas zonas. As transações de moedas no pacote IBC
precisam ser feitas pelo remetente, hub, e blockchain recebedor.

Desde a atuação do Cosmos Hub como ledger principal para todos o
sistema, a segurança do Hub é de suma importância. Enquanto cada
zona pode ser uma blockchain Tendermint que é segurada por 4((ou talvez
menos caso o consenso BFT não seja necessário), o Hub precisa ser segurado por uma descentralização
globalizada feita pelos validadores que podem evitar os mais severos tipos de
ataques, como uma partição de rede continental ou um estado-nação fazendo
ataques.

### As Zonas

Uma zona Cosmos é uma blockchain independente das trocas de mensagens IBC com o
Hub. Na perspectiva do Hub, uma zona é uma conta multi-asset dynamic-membership
multi-signature que pode enviar e receber tokens usando pacotes IBC. Como
uma conta de criptomoeda, uma zona não pode transferir mais tokens do que ela possui, mas
pode receber tokens de outras que os tem. Uma zona pode ser usada como uma
"fonte" de um ou mais tipos de tokens, garantindo o poder de aumentar o estoque desse
token.

Os atoms do Cosmos Hub podem ser estacados por validadores de uma zona conectada ao
Hub. Enquanto os ataques de gasto-duplo nesses zonas podem resultar em um core dos
atoms com o fork-responsável da Tendermint, uma zona onde +⅔ do poder de voto
são Bizantinos podem deixar o estado inválido. O Cosmos Hub não verifica ou
executa transações ocorridas em outras zonas, então essa é uma responsabilidade dos
usuários para enviar os tokes ara zonas que eles confiem. Futuramente, o sistema de
governança do Cosmos Hub irá implementar propostas para o Hub e para as falhas
das zonas. Por exemplo, um token de saída transferido para algumas (ou todas) zonas podem ser
segurados em caso de uma emergência de quebra de circuito das zonas(uma parada temporária
nas transferências dos tokens) quando um ataque é detectado.

## Comunicação Inter-blockchain (IBC)

Agora nós olhamos para como o Hub e as zonas vão se comunicar. Por exemplo, se
aqui são três blockchains, "Zona1", "Zona2", and "Hub", e nós queremos que a
"Zona1" produza um pacote destinado para a "Zona2" indo através do "Hub". Para mover um
pacote de uma blockchain para outra, uma prova é feita na
cadeia recebedora. A prova atesta esse envio publicado na cadeia de destino por uma alegação
de pacote. Para a cadeia recebedora checar essa prova, isso é possível
por um block principal de envio. Esse mecanismo é similar ao usado por
cadeias paralelas, que requerem duas cadeias interagindo uma com a outra via
transmissões bidirecionais por dados de prova-de-existência (transações).

O protocolo IBC pode naturalmente ser definido usando dois tipos de transações: uma
transação `IBCBlockCommitTx`, a qual permite uma blockchain provar para qualquer
espectador o mais recente hash-de-bloco, e uma transação `IBCPacketTx`, a qual
permite uma blockchain provar para qualquer espectador que o pacote recebido foi realmente
publicado pelo remetente, via Merkle-proof para um hash-de-bloco
recente.

Ao misturar o mecanismo ICB em duas transações separadas, nós permitimos
que o mecanismo de mercado de taxa nativa da blockchain recebedora determine quais pacotes
irão se comprometer (isto é, ser reconhecido), permitindo simultaneamente que uma
blockchain envie de quantos pacotes de saída forem permitidos.

![Figura da Zona1, Zona2, e Hub IBC sem
reconhecimento](https://raw.githubusercontent.com/gnuclear/atom-whitepaper/master/msc/ibc_without_ack.png)

No exemplo acima, para atualizar o hash de blocos da
"Zona1" no "Hub" (ou do "Hub" para a "Zona2"), uma transação `IBCBlockCommitTx`
precisa ser feita no "Hub" com o hash de bloco da "Zona1" (ou na
"Zona2" com o hash de bloco do "Hub").

_Veja [IBCBlockCommitTx](#ibcblockcommittx) e [IBCPacketTx](#ibcpacketcommit)
para mais informações sobre os 2 tipos de transação IBC._

## Casos de Uso

### Exchange Distribuídas

Da mesma forma que Bitcoin é mais seguro por ter uma distribuíção,
e replicação em massa, podemos tornar as exchanges menos vulneráveis a
Hacks internos executando-a no blockchain. Chamamos isso de exchange
distribuída.

O que a comunidade de criptomoedas chama hoje de intercâmbio descentralizado
baseado em algo chamado transações "atomic cross-chain" (AXC). Com uma transação
AXC, dois usuários em duas diferentes cadeias podem fazer duas transações
de transferências que serão feitas juntas nas duas ledgers, ou nenhuma (isto é,
Atomicamente). Por exemplo, dois usuários podem trocar bitcoins por ether (ou qualquer dois
Tokens em dois ledgers diferentes) usando transações AXC, mesmo que o Bitcoin
e o Ethereum não estão conectados entre si. O benefício de executar um
troca em transações AXC é que nenhum dos usuários precisam confiar um no outro ou
no serviço de correspondência comercial. A desvantagem é que ambas as partes precisam estar
on-line para o negócio ocorrer.

Outro tipo de intercâmbio descentralizado é um sistema de
exchange que funciona em seu próprio blockchain. Os usuários deste tipo de exchange podem
enviar uma ordem de limite e desligar o computador, e o negócio pode ser executado
sem que o usuário esteja online. O blockchain combina e completa o negócio
em nome do negociante.

Uma exchange centralizada pode criar um vasto livro de ordens de ordens e
atrair mais comerciantes. A liquidez gera mais liquidez no mundo cambial,
e assim há um forte efeito na rede (ou pelo menos efeito de vencedor-leva-mais)
no negócio de câmbio. A atual líder para troca de criptomoedas
hoje é a Poloniex com um volume de 24 milhões de dólares por dia, e em segundo lugar a
Bitfinex com um volume de US$5 milhões por dia. Dados esses fortes efeitos na rede,
é improvável que as exchanges descentralizadas baseadas no AXC ganhem volume
centrais. Para uma exchange descentralizada competir com um
exchange centralizada, seria necessário dar suporte aos livros de
ordens. Somente uma exchange distribuída em uma blockchain pode fornecer isso.

Tendermint fornece benefícios adicionais para realizar uma transação mais rápida. Com a
finalidade de dar prioridade a rapidez sem sacrificar a consistência, as zonas no Cosmos podem
finalizar transações rápidas - tanto para transações de ordem de
transferências de tokens quanto para outras zonas IBC.

Dado o estado das exchanges de criptomoedas hoje em dia, uma grande
exchange distribuída da Cosmos (aka o Cosmos DEX). A transação e
a capacidade de processamento, bem como a latência de processos, podem ser
centrais. Os comerciantes podem enviar ordens de limite que podem ser executadas
sem que ambas as partes tenham que estar online. E com Tendermint, o Cosmos Hub,
e o IBC, os comerciantes podem mover fundos dentro e fora da exchange e para outras
zonas com rapidez.

### Pegging para Outras Criptomoedas

Uma zona privilegiada pode agir como token simbolico de uma
criptomoeda. A peg é semelhante à relação entre uma
zona e o Cosmos Hub; Ambos devem manter-se atualizados com os
outros, afim de verificar provas de que os tokens passaram de um para o outro. A
Peg-zone na rede Cosmos mantém-se com o Hub, bem como o
outra cryptomoeda. A indireção através da peg-zone permite a lógica de
que o Hub permaceça simples e imutável para outras estratégias de consenso blockchain
como a mineração de prova-de-trabalho do Bitcoin.

Por exemplo, uma zona Cosmos com um conjunto validador específico, possivelmente o mesmo que
o Hub, poderia atuar como um ether-peg, onde a aplicação TMSP sobre
a zona ("peg-zone") tem mecanismos para trocar mensagens IBC com um
Peg-contract na blockchain (a "origem"). Este contrato
permite que os titulares de ether enviem ether para a zona de peg, enviando-o para
Peg-contract na Ethereum. Uma vez que o ether é recebido pelo peg-contract, o ether
não pode ser retirado a menos que um pacote IBC apropriado seja recebido pelo
Peg-contract da peg-zone. Quando uma zona recebe um pacote IBC provando
que o ether foi recebido no peg-contract para uma determinada conta Ethereum,
a conta correspondente é criada na peg-zone com esse saldo. O ether na
peg-zone ("pegged-ether") pode então ser transferido para o Hub,
e mais tarde ser destruído com uma transação que envia para um determinado
endereço de retirada no Ethereum. Um pacote IBC provando que a transação
na Peg-Zone podem ser lançados no peg-contract Ethereum para permitir que o
Ether seja retirado.

Naturalmente, o risco do contrato do pegging e um conjunto de validadores desonestos. -bizantino.
O poder de voto bizantino poderia causar um fork, retirando o ether do
peg-contract mantendo o pegged-ether na peg-zone. Na pior das hipóteses,
\+⅔ do poder de voto bizantino pode roubar o ether daqueles que o enviaram para o
peg-contract, desviando-se da pegging e da peg-zone de origem.

É possível abordar essas questões projetando o peg para ser totalmente
responsável. Por exemplo, todos os pacotes IBC, a partir do hub de
origem, poderão exigir um reconhecimento pela zona de fixação de tal forma que
as transições de estados da peg-zone podem ser desafiadas de forma eficiente e verificadas pelo
hub ou pelo peg-contract de origem. O Hub e a origem devem
permitir que os validadores da zona de fixação apresentem garantias e as transferências
contratuais devem ser atrasadas (e um prazo de vinculação de colateral suficientemente
longo) para permitir que quaisquer desafios sejam feitos por auditores independentes. Nós saímos
da concepção, da especificação e implementação deste sistema aberto como uma
futura proposta de melhoria da Cosmos, a ser aprovada pela governança do sistema do Cosmos
Hub.

Embora a atmosfera sociopolítica ainda não esteja bastante desenvolvida, podemos
aumentar o mecanismo para permitir que as zonas se liguem as moedas FIAT de um
estado-nação, formando um validador responsável estabelecido a partir de uma combinação da
moeda da nação, mais particularmente, pelos seus bancos. Claro,
precauções adicionais devem ser tomadas para aceitar apenas moedas apoiadas por
sistemas que possam reforçar a capacidade de auditoria das atividades dos bancos
e de notário grupos de grandes instituições de confiança.

Um resultado dessa integração poderia ser, por exemplo, permitir que
uma conta em um banco na zona possa mover dólares de sua conta bancária
para outras contas na zona, ou para o hub, ou inteiramente para outra zona.
Nesse sentido, o Cosmos Hub pode atuar como um canal sem
moedas e criptomoedas, removendo as barreiras que limitariam
sua interoperabilidade com o mundo dos intercâmbios.

### Ethereum Scaling

Resolver o problema de escalonamento é um problema aberto para a Ethereum. Atualmente,
os nós Ethereum processam cada transação única e também armazenam todos os estados.
[link](https://docs.google.com/presentation/d/1CjD0W4l4-CwHKUvfF5Vlps76fKLEC6pIwu1a_kC_YRQ/mobilepresent?slide=id.gd284b9333_0_28).

Desde que a Tendermint pode realizar os blocos muito mais rápido do que a prova-de-trabalho da Ethereum,
as zonas EVM alimentadas e operando pelo consenso da Tendermint
fornecem maior desempenho para blocos da blockchain Ethereum. Além disso, embora o
Cosmos Hub e o mecanismo de pacotes IBC não permitam a execução da lógica de contratos
arbitrários, podem ser usados para coordenar os movimentos Ethereum e a execução de
contratos simbólicos em diferentes zonas, fornecendo uma base para
o token ethereum através de sharding.

### Integração de Multi-Aplicação

As zonas Cosmos executam lógica de aplicação arbitrária, que é definida no início da
vida da zona e podem potencialmente ser atualizados ao longo do tempo pela governança. Essa flexibilidade
permite que as zonas Cosmos ajam como pegs para outras criptomoedas como Ethereum ou
Bitcoin, e também permite derivados desses blockchains, que utilizam a
mesma base de código, mas com um conjunto de validador diferente e distribuição inicial. Isto
permite que muitos tipos de criptomoedas existentes, como as Ethereum,
Zerocash, Bitcoin, CryptoNote e assim por diante, possam ser usados com o Tendermint Core,
que é um motor de consenso de maior desempenho, em uma rede comum, abrindo
oportunidade de interoperabilidade entre as plataformas. Além disso, como
multi-asset blockchain, uma única transação pode conter vários
onde cada entrada pode ser qualquer tipo de token, permitindo a Cosmos
ser uma plataforma para a exchange descentralizada, mesmo que as ordens sejam
para outras plataformas. Alternativamente, uma zona pode servir como um
fault-tolerant (com livros de ordens), o que pode ser uma melhoria
nas exchanges centralizadas de criptomoeda que tendem a ser invadidas com
o tempo.

As zonas também podem servir como versões bloqueadas de empresas e
sistemas, onde partes de um serviço particular da organização ou grupo de organizações
que são tradicionalmente executadas como um aplicativo TMSP
em uma certa zona, o que lhe permite receber a segurança e a interoperabilidade da
rede pública Cosmos sem sacrificar o controle sobre o serviço subjacente.
Assim, a Cosmos pode oferecer o melhor da tecnologia blockchain para ambos os mundos e
para as organizações, que se recusam a deixar completamente o controle
para um distribuidor terceirizado.

### Redução de partição de rede

Alguns afirmam que um grande problema da coerência-favorecendo algoritmos de consenso
como o Tendermint é que qualquer partição de rede que faz com que não haja uma única
partição com +⅔ de poder de votação (por exemplo, ⅓+ ficando offline) irá parar o consenso
completamente. A arquitetura Cosmos pode ajudar a mitigar esse problema usando umas
zonas regionais autônomas, onde o poder de voto para cada zona é
distribuído com base em uma região geográfica comum. Por exemplo, um
parâmetro pode ser para cidades individuais, ou regiões, para operar suas próprias zonas
de partilha com um centro em comum (por exemplo, o Cosmos Hub), permitindo que a
o hub possa parar devido a uma paralisação de rede temporária.
Observe que isso permite uma geologia real, política e rede-topológica,
que são recursos a serem considerados no projeto de sistemas robustos federados de fault-tolerant.

### Sistema de Resolução de Nomes Federados

NameCoin foi uma das primeiras blockchains a tentar resolver o
problema de resolução de nomes através de uma adaptação da blockchain do Bitcoin. Infelizmente
têm ocorrido várias questões com esta abordagem.

Com a Namecoin, podemos verificar que, por exemplo, o nome <em>@satoshi</em> foi registrado como
particular, em algum momento do passado, mas não saberíamos se
a chave pública tinha sido atualizada recentemente, a menos que baixassemos todos os blocos
desde a última atualização desse nome. Isto é devido as limitações do modelo de
Merkle-ization de UTXO do Bitcoin, onde somente as transações (não
mutáveis) são Merkle-ized no hash do bloco. Isso nos permite
provar a existência, mas não a não-existência de atualizações posteriores a um nome. Assim, nós
não podemos saber com certeza o valor mais recente de um nome sem confiar em um
nó, ou recorrer a gastos significativos na largura de banda, baixando o
Blockchain.

Mesmo se uma árvore de pesquisa Merkle-ized for implementada na NameCoin, sua dependência
sobre a prova-de-trabalho torna a verificação do cliente light problemática. Os clientes light devem
baixar uma cópia completa dos cabeçalhos para todos os blocos em toda a blockchain
(ou pelo menos todos os cabeçalhos desde a última atualização de um nome). Isso significa que
os requisitos de largura de banda crescem linearmente com a o passar do tempo [\[21\]][21].
Além disso, as mudanças de nome em um bloco de prova-de-trabalho requerem
a confirmação do trabalho, o que pode levar até uma hora
no Bitcoin.

Com Tendermint, tudo o que precisamos é o hash de bloco mais recente assinado por um quorum de
validadores (por poder de voto), e uma prova Merkle para o valor atual associado
com o nome. Isto torna possível ter uma solução sucinta, rápida e segura
para a verificação de valores de nome no cliente light.

Na Cosmos, podemos aplicar este conceito e estendê-lo ainda mais. Cada
zona de registro de nomes na Cosmos pode ter um domínio de nível superior (TLD)
associado, como o ".com" ou ".org", e cada zona de registro de nome pode ter
suas próprias regras de governança e registro.

## Emissão e Incentivos

### O Token Atom

Enquanto o Cosmos Hub é um ledger de distribuíção multi-asset, há um token nativo
especial chamado _atom_. Os atoms são o únicos símbolos do Cosmos
Hub. Os atoms são uma licença para o titular votar, validar ou delegar
validadores. Como o ether da Ethereum, os atoms também podem ser usados para
reduzir o spam. Atoms inflacionários adicionais e as taxas do bloco de transação
são recompensadas pelos validadores e delegados que
o validarão.

A transação `BurnAtomTx` pode ser usada para cobrir proporcionalmente a quantidade
de tokens reservados para a pool.

#### Levantamento de Fundos

A distribuição inicial dos tokens atom e validadores na Genesis vão para os
doadores do Levantamento de Fundos da Cosmos (75%), doadores pesados (5%), Fundação da Rede
Cosmos (10%), e a ALL IN BITS, Inc (10%). A partir da Genesis em diante, 1/3 da
quantidade total de atoms será recompensada aos validadores e delegados durante
todo o ano.

Veja o [Plano Cosmos](https://github.com/cosmos/cosmos/blob/master/PLAN.md)
para detalhes adicionais.

#### Investindo

Para evitar que o levantamento de fundos atraia especuladores de curto prazo apenas interessados
em esquemas de pump and dump, os atoms da Genesis não serão transferíveis até
eles tenham investido. Cada conta irá adquirir atoms durante um período de 2 anos com
taxa constante a cada hora, determinada pelo número total de atoms da Genesis/(2*
365 * 24) horas. Os atoms ganhos pela recompensa do bloco são pré-investidos,
e podem ser transferidos imediatamente, de modo que os validadores e os delegados ligados possam ganhar
mais da metade de seus atoms da Genesis após o primeiro ano.

### Limitações do Número de Validadores

Diferentemente do Bitcoin ou de outros blockchains de prova-de-trabalho, o blockchain Tendermint será
mais lento com mais validadores devido ao aumento da complexidade da comunicação.
Felizmente, podemos oferecer suporte a validadores suficientes para a
distribuição na Blockchain com tempos de confirmação de transação muito mais rápidos e, através de
largura de banda, armazenamento e aumento da capacidade de computação paralela, seremos capazes de
ter mais validadores no futuro.

No dia da Genesis, o número máximo de validadores será definido como 100,
o número aumentará a uma taxa de 13% durante 10 anos até atingir a marca de 300
Validadores.

    Ano 0: 100
    Ano 1: 113
    Ano 2: 127
    Ano 3: 144
    Ano 4: 163
    Ano 5: 184
    Ano 6: 208
    Ano 7: 235
    Ano 8: 265
    Ano 9: 300
    Ano 10: 300
    ...

### Tornando-se um Validador depois do dia da Genesis

Os titulares de atoms que ainda não são capazes de se tornarem validadores assinados e
submeter uma transação `BondTx`. A quantidade de atoms fornecida como garantia
deve ser diferente de zero. Qualquer pessoa pode se tornar um validador a qualquer momento, exceto quando o
tamanho do conjunto de validadores atual é maior que o número máximo de
validadores permitidos. Nesse caso, a transação só é válida se o montante
de atoms é maior do que a quantidade de atoms efetivos mantidos pelo menor
validador, onde atoms eficazes incluem atoms delegados. Quando um novo validador
substitui um validador existente de tal forma, o validador existente torna-se
inativo e todos os atoms e atoms delegados entram no estado de unbonding.

### Penalidades para Validadores

Deve haver alguma penalidade imposta aos validadores por qualquer desvio intencional
ou não intencional do protocolo sancionado. Algumas evidências são imediatamente admissíveis,
como um double-sign na mesma altura e volta, ou uma violação de "prevote-the-lock"
(uma regra do protocolo de consenso Tendermint). Tais evidências resultarão em que o
validador perca sua boa reputação e seus átomos ligados, bem como sua proporção de tokens
na pool reserva - coletivamente chamados de "stake" - serão cortados.

Às vezes, os validadores não estarão disponíveis, devido a interrupções na rede regional,
falha de energia ou outros motivos. Se, em qualquer ponto nos blocos `ValidatorTimeoutWindow`
anteriores, o voto de validação de um validador não estiver incluído na cadeia de
blocos mais do que `ValidatorTimeoutMaxAbsent` vezes, esse validador ficará inativo e
perderá `ValidatorTimeoutPenalty` (PADRÃO DE 1%) de sua participação.

Alguns comportamentos "maliciosos" não produzem provas obviamente discerníveis sobre
a blockchain. Nesses casos, os validadores podem coordenar fora da banda para forçar
o tempo limite desses validadores maliciosos, se houver um consenso majoritário.

Em situações em que o Cosmos Hub parar devido a uma coalizão de ⅓+ de poder de voto
offline, ou em situações onde uma coalizão de ⅓+ de poder de voto censurar evidências de
comportamento malicioso entrando na blockchain, o hub deve recuperar com um hard-fork
de proposta reorganizacional. (Link to "Forks and Censorship Attacks").

### Taxas de Transação

Os validadores do Cosmos Hub podem aceitar qualquer tipo de token ou combinação
de tipos como taxas para processar uma transação. Cada validador pode fixar subjetivamente a
taxa de câmbio que quiser e escolher as transações que desejar, desde que o `BlockGasLimit`
não seja excedido. As taxas cobradas, menos quaisquer impostos especificados abaixo,
são redistribuídas aos stakeholders ligados em proporção aos seus átomos ligados, cada `ValidatorPayoutPeriod` (PADRÃO DE 1 hora).

Das taxas de transação cobradas, `ReserveTax` (PADRÃO DE 2%) irá para a pool reserva
para aumentar a pool reserva e aumentar a segurança e o valor da rede Cosmos. Além disso, um
`CommonsTax` (PADRÃO DE 3%) irá para o financiamento de bens comuns. Estes fundos vão para o
`CustodianAddress` para ser distribuído de acordo com as decisões tomadas pelo sistema de governança.

Os titulares de átomos que delegam o seu poder de voto a outros validadores pagam uma comissão
ao validador delegado. A comissão pode ser definida por cada validador.

### Incentivando Hackers

A segurança do Cosmos Hub é uma função da segurança dos validadores subjacentes e da escolha
da delegação pelos delegados. A fim de incentivar a descoberta e notificação precoce de vulnerabilidades
encontradas, o Cosmos Hub incentiva os hackers a publicar exploits bem sucedidos através de uma transação
`ReportHackTx` que diz," Este validador foi hackeado. Por favor, envie recompensa para este endereço".
Depois de tal exploração, o validador e os delegados ficarão inativos, `HackPunishmentRatio` (PADRÃO DE 5%)
dos átomos de todos serão cortados, e`HackRewardRatio` (PADRÃO DE 5%) dos átomos de todos
serão recompensado com o endereço de recompensa do hacker. O validador deve recuperar os átomos
restantes usando sua chave de backup.

Para evitar que esse recurso seja abusado para transferir átomos não invadidos,
a porção de átomos adquirido vs relativo de validadores e delegados antes e depois do `ReportHackTx`
permanecerá o mesmo, e o bounty do hacker irá incluir alguns átomos relativos, se houver.

### Específicação de Governança

O Cosmos Hub é operado por uma organização distribuída que requer um mecanismo de
governança bem definido para coordenar várias mudanças na blockchain, como parâmetros
variáveis do sistema, bem como atualizações de software e emendas constitucionais.

Todos os validadores são responsáveis por votar em todas as propostas.
Não votar em uma proposta em tempo hábil resultará na desativação automática do
validador por um período de tempo denominado `AbsenteeismPenaltyPeriod` (PADRÃO DE 1 semana).

Os delegados herdam automaticamente o voto do validador delegado.
Este voto pode ser anulado manualmente. Os átomos não ligados obtêm nenhum voto.

Cada proposta requer um depósito de tokens de `MinimumProposalDeposit`,
que pode ser uma combinação de um ou mais tokens incluindo átomos.
Para cada proposta, os eleitores podem votar para receber o depósito.
Se mais da metade dos eleitores optarem por receber o depósito (por exemplo, porque a proposta era spam),
o depósito vai para a pool reserva, exceto os átomos que são queimados.

Para cada proposta, os eleitores podem votar nas seguintes opições:

- Sim
- Com Certeza
- Não
- Nunca
- Abstenção

É necessário uma maioria estrita de votos Yea(Sim) ou YeaWithForce(Com certeza)
(ou votos Nay(Não) ou NayWithForce(Nunca)) para que a proposta seja decidida como aceita
(ou decidida como falha), mas 1/3+ pode vetar a decisão da maioria votando "Com certeza".
Quando uma maioria estrita é vetada, todos são punidos com a perda de `VetoPenaltyFeeBlocks`
(PADRÃO DE no valor de um dia de blocos) de taxas (exceto os impostos que não serão afetados),
e a parte que vetou a decisão da maioria será adicionalmente punida com a perda de `VetoPenaltyAtoms`
(PADRÃO DE 0.1%) de seus átomos.

### Parâmetro de Mudança de Proposta

Qualquer um dos parâmetros aqui definidos pode ser alterado com a aceitação
de um `ParameterChangeProposal`.

### Texto da Proposta

Todas as outras propostas, como uma proposta de atualização do protocolo, serão coordenadas através do genérico `TextProposal`.

## Roteiro

Veja [o Plano Cosmos](https://github.com/cosmos/cosmos/blob/master/PLAN.md).

## Trabalho Relacionado

Houve muitas inovações no consenso da blockchain e na escalabilidade nos últimos dois anos.
Esta seção fornece um breve levantamento de um seleto número das mais importantes.

### Sistemas de Consenso

#### Classic Byzantine Fault Tolerance

Consenso na presença de participantes maliciosos é um problema que remonta ao início dos anos 1980,
quando Leslie Lamport cunhou a frase "falha bizantina" para se referir ao comportamento do processo
arbitrário que se desvia do comportamento pretendido, que contraste com uma "falha acidental",
em que um processo simplesmente falha. Soluções iniciais foram descobertas para redes síncronas onde
há um limite superior na latência da mensagem, embora o uso prático fosse limitado a ambientes altamente controlados,
como controladores de avião e datacenters sincronizados via relógios atômicos.
Não foi até o final dos anos 90 que a Practical Byzantine Fault Tolerance (PBFT) foi introduzida como
um eficiente algoritmo de consenso parcialmente síncrono capaz de tolerar até ⅓ de processos
comportando-se arbitrariamente. PBFT tornou-se o algoritmo padrão, gerando muitas variações,
incluindo mais recentemente uma criada pela IBM como parte de sua contribuição para a Hyperledger.

O principal benefício do consenso Tendermint sobre PBFT é que o Tendermint tem uma estrutura
subjacente melhorada e simplificada, um dos quais é um resultado de adotar o paradigma blockchain.
Blocos Tendermint devem confirmar em ordem, o que evita a complexidade e sobrecarga de comunicação
associada a alteração de visão do PBFT's. No Cosmos e muitas outras criptomoedas,
não há necessidade de permitir o bloco <em>N+i</em> onde <em>i >= 1</em> se confirmar,
quando o próprio bloco <em>N</em> ainda não se confirmou. Se a largura de banda é a razão
pela qual o bloco <em>N</em> não se confirmou em uma zona do Cosmos, então isso não ajuda
a usar os votos de compartilhamento de largura de banda para blocos <em>N+i</em>.
Se uma partição de rede ou nós offline for a razão pela qual o bloco <em>N</em> não foi confirmado,
<em>N+i</em> não se comprometerá de qualquer maneira.

Além disso, o lote de transações em blocos permite que o Merkle-hashing regule o estado da aplicação,
ao invés de resumos periódicos com esquemas de pontos de verificação como PBFT faz.
Isso permite confirmações de transações mais rápidas para clientes leves e uma comunicação mais rápida entre a blockchain.

Tendermint Core também inclui muitas otimizações e recursos que vão acima e além do que é especificado no PBFT.
Por exemplo, os blocos propostos pelos validadores são divididos em partes,
Merkleized e inútilizados de tal forma que melhora o desempenho da transmissão
(ver LibSwift [\[19\]][19] para inspiração). Além disso, Tendermint Core não faz qualquer suposição sobre
a conectividade ponto-a-ponto, e funciona durante o tempo que a rede P2P está fracamente conectada.

#### Participação delegada do BitShares

Apesar de não serem os primeiros a implementar a prova-de-participação (Proof-of-Stake - PoS),
o BitShares [\[12\]][12] contribuiu consideravelmente para a pesquisa e adoção das blockchains que usam o PoS,
particularmente aqueles conhecidos como PoS "delegados". No BitShares, as partes interessadas elegem "testemunhas",
responsáveis por ordenar e confirmar transações e "delegados", responsáveis pela coordenação
de atualizações de software e alterações de parâmetros. Embora o BitShares atinja alto desempenho
(100k tx/s, 1s de latência) em condições ideais, ele está sujeito a ataques de duplo gasto por testemunhas
maliciosas que "forkem" a blockchain sem sofrer uma punição econômica explícita - ele sofre do problema
"nada a perder". O BitShares tenta suavizar o problema permitindo que as transações se refiram a
blocos-hashes recentes. Além disso, as partes interessadas podem remover ou substituir
testemunhas de má conduta diariamente, embora isso não faça nada para punir
explicitamente os ataques bem sucedidos de duplo gasto.

#### Stellar

Baseando-se em uma abordagem pioneira da Ripple, a Stellar [\[13\]][13] refinou um modelo do
Federated Byzantine Agreement em que os processos que participam do consenso não constituem
um conjunto fixo e globalmente conhecido. Em vez disso, cada nó de processo codifica uma ou mais
"fatias de quórum", cada uma constituindo um conjunto de processos confiáveis. Um "quórum" na
Stellar é definido como um conjunto de nós que contêm pelo menos uma fatia de quórum para cada
nó no conjunto, de modo que o acordo possa ser alcançado.

A segurança do mecanismo Stellar baseia-se no pressuposto de que a intersecção de _qualquer_ dois
quóruns é não-vazia, enquanto a disponibilidade de um nó requer pelo menos uma das suas fatias de
quórum para consistir inteiramente de nós corretos, criando um troca externa entre o uso de grandes
ou pequenas fatias-quórum que podem ser difíceis de equilíbrar sem impor pressupostos significativos
sobre a confiança. Em última análise, os nós precisam, de alguma forma, escolher fatias de quórum adequadas
para que haja tolerância suficiente a falhas (ou qualquer "nó intacto" em geral, do qual muitos dos
resultados do trabalho dependem) e a única estratégia fornecida para garantir tal configuração é
hierárquica e similar ao Border Gateway Protocol (BGP), usado por ISPs de primeira linha na
internet para estabelecer tabelas de roteamento globais e usado pelos navegadores para gerenciar
certificados TLS; Ambos notórios por sua insegurança.

A crítica sobre papel da Stellar nos sistemas PoS baseados em Tendermint é atenuada pela estratégia
de token descrita aqui, em que um novo tipo de token chamado _atom_ é emitido para representar
reivindicações para futuras porções de taxas e recompensas. A vantagem do PoS baseado em Tendermint,
portanto, é a sua relativa simplicidade, ao mesmo tempo que oferece garantias de segurança suficientes e prováveis.

#### BitcoinNG

O BitcoinNG é uma proposta de melhoria do Bitcoin que permitiria formas de escalabilidade vertical,
como o aumento do tamanho do bloco, sem as conseqüências econômicas negativas normalmente associadas a tal mudança,
como o impacto desproporcionalmente grande sobre os pequenos mineradores. Esta melhoria é conseguida separando
a eleição do líder da transmissão da transação: os líderes são eleitos pela primeira vez
por prova de trabalho(PoW) em "microblocos", e então são capazes de transmitir transações a
serem confirmadas até que um novo microbloco seja encontrado. Isso reduz os requisitos
de largura de banda necessários para vencer a corrida PoW, permitindo que os pequenos
mineiros possam competir mais justamente, e permitindo que as transações sejam confirmadas
com mais regularidade pelo último minerador para encontrar um micro-bloco.

#### Casper

Casper [\[16\]][16] é uma proposta de algoritmo de consenso PoS para o Ethereum.
Seu modo principal de operação é "consenso-por-aposta". Ao permitir que os validadores apostem
iterativamente em qual bloco eles acreditam que será confirmado na blockchain com base nas
outras apostas que eles têm visto até agora, a finalidade pode ser alcançada eventualmente.
[link](https://blog.ethereum.org/2015/12/28/understanding-serenity-part-2-casper/). Esta é uma área ativa
de pesquisa da equipe de Casper. O desafio está na construção de um mecanismo de apostas que pode ser
comprovado como uma estratégia evolutivamente estável. O principal benefício da Casper em relação à
Tendermint pode ser a oferta de "disponibilidade sobre a consistência" - consenso não requer
um quórum +⅔ de poder de voto - talvez ao custo de velocidade de confirmação ou complexidade de implementação.

### Escala Horizontal

#### Protocolo Interledger

O Protocolo Interledger [\[14\]][14] não é estritamente uma solução de escalabilidade.
Ele fornece uma interoperabilidade ad hoc entre diferentes sistemas de ledger através de uma rede
de relações bilaterais livremente acopladas. Tal como a Lightning Network, a finalidade do
ILP é facilitar pagamentos, mas focaliza especificamente pagamentos em diferentes tipos de ledger,
estendendo o mecanismo de transações atômicas para incluir não apenas hash-locks, mas também um
quórum de notários (chamado de Atomic Transport Protocol). O último mecanismo para reforçar a
atomicidade em transacções entre-ledger é semelhante ao mecanismo SPV do cliente leve do Tendermint,
então uma ilustração da distinção entre ILP e Cosmos/IBC é garantida, e fornecida abaixo.

1.  Os notários de um conector em ILP não suportam mudanças de consentimento, e não permitem uma
    pesagem flexível entre notários. Por outro lado, o IBC é projetado especificamente para blockchains,
    onde os validadores podem ter diferentes pesos, e onde o consentimento pode mudar ao longo da cadeia de blocos.

2.  Como na Lightning Network, o receptor do pagamento em ILP deve estar on-line para enviar
    uma confirmação de volta ao remetente. Em uma transferência de token sobre IBC, o conjunto
    de validadores da blockchain do receptor é responsável por fornecer a confirmação, não o usuário receptor.

3.  A diferença mais notável é que os conectores do ILP não são responsáveis ou mantêm o estado
    autoritário sobre os pagamentos, enquanto que no Cosmos, os validadores de um hub são a autoridade
    do estado das transferências de tokens do IBC, bem como a autoridade da quantidade de tokens
    mantidos por cada zona (mas não a quantidade de tokens mantidos por cada conta dentro de uma zona).
    Esta é a inovação fundamental que permite a tranferência assimétrica segura de tokens de zona para
    zona; O conector analógico do ILP no Cosmos é uma persistente e maximamente segura ledger de blockchain, o Cosmos Hub.

4.  Os pagamentos entre contas no ILP precisam ser suportados por uma ordem de compra/venda, uma
    vez que não há transferência assimétrica de moedas de um ledger para outro, apenas a transferência
    de valor ou equivalentes de mercado.

#### Sidechains

Sidechains [\[15\]][15] são um mecanismo proposto para dimensionar a rede Bitcoin através de
blockchains alternativas que são "atreladas" para a blockchain do Bitcoin. As Sidechains
permitem que bitcoins se movam efetivamente da blockchain do Bitcoin para a sidechain e retornarem,
e permitem a experimentação em novos recursos na sidechain. Como no Cosmos Hub, a sidechain e
Bitcoin servem como clientes leves uns dos outros, usando provas SPV para determinar quando as moedas
devem ser transferidas para a cadeia lateral e retornarem. Claro, como o Bitcoin usa PoW, sidechains
centradas em torno do Bitcoin sofrem dos muitos problemas e riscos do PoW como um mecanismo de consenso.
Além disso, esta é uma solução Bitcoin-maximalista que não suporta nativamente uma variedade de tokens e
topologia de rede entre-zona como o Cosmos faz. Dito isto, o mecanismo de núcleo bidirecional atrelado é,
em princípio, o mesmo que o empregado pela rede Cosmos.

#### Esforços de Escalabilidade do Ethereum

Ethereum está atualmente pesquisando uma série de estratégias diferentes para fragmentar o
estado da blockchain do Ethereum para atender às necessidades de escalabilidade.
Esses esforços têm como objetivo manter a camada de abstração oferecida pela atual
Ethereum Virtual Machine através do espaço de estado compartilhado. Vários esforços de
pesquisa estão em andamento neste momento. [\[18\]][18][\[22\]][22]

##### Cosmos vs Ethereum 2.0 Mauve

Cosmos e Ethereum 2.0 Mauve [\[22\]][22] tem diferentes objetivos de projeto.

- Cosmos é especificamente sobre tokens. Malva é sobre escalonamento de computação geral.
- O Cosmos não está ligado ao EVM, por isso mesmo VMs diferentes podem interoperar.
- Cosmos permite que o criador da zona determine quem valida a zona.
- Qualquer pessoa pode iniciar uma nova zona no Cosmos (a menos que a governança decida o contrário).
- O hub isola falhas de zonas de modo que tokens invariantes sejam preservados.

### Escala Geral

#### Lightning Network

A Lightning Network é uma proposta de rede de transferência de token operando em uma camada acima
da blockchain do Bitcoin (e outras blockchains públicas), permitindo a melhoria de muitas ordens
de magnitude no processamento de transações movendo a maioria das transações fora da ledger de consenso
para o chamado "Canais de pagamento". Isso é possível graças a scripts de criptomoedas em cadeia,
que permitem que as partes entrem em contratos estatais bilaterais onde o estado pode ser atualizado
compartilhando assinaturas digitais, e os contratos podem ser fechados definitivamente publicando
evidências na blockchain, um mecanismo primeiramente popularizado por trocas atômicas de
cross-chains(cadeias cruzadas). Ao abrir canais de pagamento com muitas partes, os participantes
da Lightning Network podem se tornar pontos focais para encaminhar os pagamentos de outros,
levando a uma rede de canais de pagamento totalmente conectada, ao custo do capital estar ligado aos canais de pagamento.

Enquanto a Lightning Network também pode facilmente se estender através de várias blockchains independentes
para permitir a transferência de _value_ através de um mercado de câmbio, não pode ser usado para
transferir assimetricamente _tokens_ de uma blockchain para outra. O principal benefício da rede Cosmos
descrita aqui é permitir tais transferências diretas de tokens. Dito isto, esperamos que os canais de
pagamento e a Lightning Network sejam amplamente adotados juntamente com nosso mecanismo de transferência
de token, por razões de economia de custos e privacidade.

#### Segregated Witness

Segregated Witness é uma proposta de melhoria do Bitcoin
[link](https://github.com/bitcoin/bips/blob/master/bip-0141.mediawiki) que visa aumentar em 2X ou 3X a
taxa de transferência por bloco, ao mesmo tempo que faz a sincronização de blocos ser mais rapida para
novos nós. O brilho desta solução é de como ele funciona dentro das limitações do protocolo atual do Bitcoin
e permite uma atualização de soft-fork (ou seja, os clientes com versões mais antigas do software
continuarão funcionando após a atualização). O Tendermint, sendo um novo protocolo, não tem restrições
de projeto, por isso tem prioridades diferentes de escalonamento. Sobretudo, o Tendermint usa um algoritmo
de rodízio BFT baseado em assinaturas criptográficas em vez de mineração, o que trivialmente permite escalonamento
horizontal através de múltiplas blockchains paralelas, enquanto que os regulares e mais frequentes blocos confirmam
a escala vertical também.

<hr/>

## Apêndice

### Responsabilidade de Fork

Um protocolo de consenso bem projetado deve fornecer algumas garantias no caso da capacidade de
tolerância ser excedida e o consenso falhar. Isto é especialmente necessário nos sistemas econômicos,
onde o comportamento Bizantino pode ter recompensa financeira substancial. A
garantia maisimportante é uma forma de _fork-accountability_, onde os processos que
fizeram com que o consenso falhasse (ou seja, clientes do protocolo
motivados para aceitar valores diferentes - um fork) podem ser identificados e punidos de acordo com as
regras do protocolo , Ou, possivelmente, o sistema jurídico. Quando o sistema jurídico não é confiável
ou é excessivamente caro para suplicar, os validadores podem ser forçados a fazerem depósitos de segurança
para participar, e esses depósitos podem ser revogados ou cortados, quando um comportamento malicioso é detectado [\[10\]][10].

Observe que isso é diferente do Bitcoin, onde o fork é uma ocorrência regular devido à assincronia de
rede e à natureza probabilística de encontrar colisões de hash parciais. Uma vez que, em muitos casos,
um fork malicioso é indistinguível de um fork devido à assincronia, o Bitcoin não pode implementar de
forma confiável a responsabilidade de um fork, com exceção do custo implícito pago por mineradores que
tem a oportunidade de minerarem um bloco órfão.

### Consenso Tendermint

Chamamos as fases de votação de _PreVote_ e _PreCommit_. Um voto pode ser para um bloco em particular ou
para _Nil_. Chamamos uma coleção de +⅔ PreVotes para um único bloco na mesma rodada de um _Polka_, e uma
coleção de +⅔ PreCommits para um único bloco na mesma rodada de um _Commit_. Se +⅔ PreCommit para Nil na
mesma rodada, eles passam para a próxima rodada.

Observe que o determinismo estrito no protocolo incorre em uma suposição de sincronia fraca, pois os líderes
com falhas devem ser detectados e ignorados. Assim, os validadores aguardam algum tempo, _TimeoutPropose_,
antes de Prevote Nil, e o valor de TimeoutPropose aumenta a cada rodada. A progressão através do
resto de uma rodada é totalmente assincrôna, onde o progresso é feito somente quando um validador
ouve de +⅔ da rede. Na prática, seria necessário um adversário extremamente forte para impedir
indefinidamente a suposição de sincronia fraca (fazendo com que o consenso deixasse de confirmar um bloco),
e isso pode ser ainda mais difícil usando valores randomizados de TimeoutPropose em cada validador.

Um conjunto adicional de restrições, ou Locking Rules(Regras de bloqueio), garante que a rede acabará
por confirmar apenas um bloco em cada altura. Qualquer tentativa maliciosa de confirmar de causar um
bloco a ser confirmado a uma determinada altura pode ser identificada. Primeiro, um PreCommit para um
bloco deve vir com justificação, na forma de um Polka para esse bloco. Se o validador já tiver PreCommit
um bloco na rodada <em>R*1</em>, nós dizemos que eles estão \_locked* nesse bloco, e o Polka usado
para justificar o novo PreCommit na rodada <em>R_2</em> deve vir de uma rodada <em>R_polka</em>
onde <em>R_1 &lt; R_polka &lt;= R_2</em>. Em segundo lugar, os validadores devem propor e/ou pré-votar
o bloco que eles estão travados. Juntas, essas condições garantem que um validador não PreCommit
sem evidência suficiente como justificativa, e que os validadores que já têm PreCommit não podem
contribuir para a evidência de PreCommit algo mais. Isso garante a segurança e a vivacidade do algoritmo de consenso.

Os detalhes completos do protocolo são descritos
[aqui](https://github.com/tendermint/tendermint/wiki/Byzantine-Consensus-Algorithm).

### Clientes Leves do Tendermint

A necessidade de sincronizar todos os cabeçalhos de bloco é eliminada no Tendermint-PoS, como por exemplo
a existência de uma cadeia alternativa (um fork) significando que ⅓+ do stake ligado pode ser reduzido.
Naturalmente, a partir que dividir requer que _someone_ compartilhe evidência de um fork, clientes leves
devem armazenar qualquer bloco-hash comprometido que eles vêem. Além disso, os clientes leves podem
periodicamente ficarem sincronizados com as alterações no conjunto de validadores, para evitar
[ataques de longo alcance](#preventing-long-range-attacks) (mas outras soluções são possíveis).

Em espírito semelhante do Ethereum, o Tendermint permite que os aplicativos incorporem um hash de raiz
Merkle global em cada bloco, permitindo verifícações fáceis de consultas de estado para fins como saldos
de contas, o valor armazenado em um contrato ou a existência de saída de uma transação não gasta,
dependendo da natureza da aplicação.

### Prevenção de ataques de longo alcance

Assumindo uma coleção suficientemente elástica de redes de difusão e um conjunto de validador
estático, qualquer fork na blockchain pode ser detectado e os depósitos dos validadores ofensivos cortados.
Esta inovação, sugerida pela primeira vez por Vitalik Buterin no início de 2014, resolve o problema do "nada a perder" de outras
criptomoedas de PoW (ver [Trabalho Relacionado](#related-work)). No entanto, uma vez que os conjuntos de
validadores devem ser capazes de mudar, durante um longo período de tempo, os validadores originais podem
tornar-se não ligados e, portanto, seriam livres para criar uma nova cadeia a partir do bloco gênese,
não incorrendo nenhum custo, visto que eles não tem depósitos trancados. Este ataque veio a ser conhecido
como Ataque de Longo Alcance (Long Range Attack - LRA), em contraste com um Ataque de Curto Alcance,
onde os validadores que estão atualmente ligados causam um fork e são, portanto, puníveis
(assumindo um algoritimo BFT de fork-responsável como o consenso Tendermint).
Ataques de longo alcance são muitas vezes pensados para serem um golpe crítico para o PoW.

Felizmente, o LRA pode ser atenuado da seguinte forma. Em primeiro lugar, para que um validador se
desatar (assim recuperando seu depósito colateral e não mais ganhando taxas para participar no consenso),
o depósito deve ser tornado intransferível por um período de tempo conhecido como o "unbonding period"
(período de desatamento), que pode ser na ordem de semanas ou meses. Em segundo lugar, para um cliente
leve ser seguro, a primeira vez que ele se conecta à rede, ele deve verificar um hash de bloqueio recente
contra uma fonte confiável ou, preferencialmente, várias fontes. Esta condição é por vezes referida como
"subjetividade fraca". Finalmente, para permanecer seguro, ele deve sincronizar com o mais recente
validador definido, pelo menos, tão frequentemente quanto a duração do período de desatamento.
Isso garante que o cliente leve saiba sobre as alterações no conjunto de validação definido antes de
um validador não ter mais o seu capital ligado e, portanto, não mais em jogo, o que permitiria enganar
o cliente, executando um ataque de longo alcance, criando novos blocos re-começando em uma altura
a qual foi ligado (assumindo que tem controle de muitas das primeiras chaves privadas).

Note que superar o LRA desta forma requer uma revisão do modelo de segurança original do PoW. No PoW,
presume-se que um cliente leve pode sincronizar com a altura atual do bloco gênese confiável a qualquer
momento simplesmente processando o PoW em cada cabeçalho de bloco. Para superar o LRA, entretanto,
exigimos que um cliente leve entre em linha com alguma regularidade para rastrear mudanças no conjunto
de validadores e que, na primeira vez em que eles fiquem on-line, eles devem ser particularmente cuidadosos
para autenticar o que ouvem da rede contra fontes confiáveis . Naturalmente, este último requisito é
semelhante ao do Bitcoin, onde o protocolo e o software também devem ser obtidos a partir de uma fonte confiável.

O método acima para prevenir LRA é bem adequado para validadores e nós completos de uma blockchain alimentada
por Tendermint porque estes nós são destinados a permanecerem conectados à rede. O método também é adequado
para clientes leves que podem ser esperados para sincronizar com a rede com freqüência. No entanto, para
os clientes leves que não se espera ter acesso frequente à Internet ou à rede da blockchain, ainda pode
ser utilizada outra solução para superar o LRA. Os detentores de tokens não validadores podem publicar
os seus tokens como colaterais com um período de não ligação muito longo (por exemplo, muito mais longo
do que o período de não ligação para validadores) e servir clientes leves com um método secundário de
atestar a validade dos blocos atuais e hashes de blocos passados. Embora esses tokens não contam para a
segurança do consenso da blockchain, eles podem fornecer fortes garantias para clientes leves. Se a
consulta histórica de hash de blocos fosse suportada no Ethereum, qualquer pessoa poderia vincular
seus tokens em um contrato inteligente projetado especialmente para isso e fornecer serviços de
comprovação de pagamentos, efetivamente criando um mercado para a segurança contra LRA de cliente leve.

### Superando Forks e Ataques de Censura

Devido à definição de uma confimação de bloco, qualquer coalizão de poder de voto ⅓+ pode interromper a
blockchain ficando off-line ou não transmitir os seus votos. Tal coalizão também pode censurar transações
particulares rejeitando blocos que incluem essas transações, embora isso resultaria em uma proporção
significativa de propostas de blocos a serem rejeitadas, o que iria retardar a taxa de blocos
confirmados da blockchain, reduzindo sua utilidade e valor. A coalizão mal-intencionada também pode transmitir
votos em um fio de modo a triturar os blocos confirmados da blockchain para quase parar, ou se envolver em
qualquer combinação desses ataques. Finalmente, isso pode fazer com que a cadeia de blocos "forke" (bifurque),
por dupla assinatura ou violação as regras de bloqueio.

Se um adversário globalmente ativo também estivesse envolvido, poderia dividir a rede de tal maneira que
possa parecer que o subconjunto errado de validadores era responsável pela desaceleração. Esta não é apenas
uma limitação do Tendermint, mas sim uma limitação de todos os protocolos de consenso cuja
rede é potencialmente controlada por um adversário ativo.

Para estes tipos de ataques, um subconjunto de validadores deve coordenar através de meios externos
para assinar um proposta de reorganização que escolhe um fork (e qualquer prova disso) e o
subconjunto inicial de validadores com suas assinaturas. Os validadores que assinam tal
proposta de reorganização deixam seu colateral em todos os outros forks. Os clientes
devem verificar as assinaturas na proposta de reorganização, verificar qualquer
evidência e fazer um julgamento ou solicitar ao usuário final uma decisão. Por exemplo,
uma carteira para celular um aplicativo que pode alertar o usuário com um aviso de segurança,
enquanto um refrigerador pode aceitar qualquer proposta de reorganização assinada por
\+½ dos validadores originais por poder de voto.

Nenhum algoritmo não-sincrônico tolerante a falhas Bizantino pode chegar a um consenso quando ⅓+
de poder de voto for desonesto, mas um fork supõe que ⅓+ do poder de voto já foram desonestos por
dupla assinatura ou bloqueio de mudança sem justificativa. Portanto, assinar a proposta de
reorganização é um problema de coordenação que não pode ser resolvido por qualquer protocolo
não-sincronico (isto é, automaticamente e sem fazer suposições sobre a confiabilidade da rede subjacente).
Por enquanto, deixamos o problema da coordenação da proposta de reorganização para a coordenação
humana através do consenso social na mídia na internet. Os validadores devem ter cuidado para garantir
que não haja partições de rede remanescentes antes de assinar uma proposta de reorganização,
para evitar situações em que duas propostas de reorganização em conflito sejam assinadas.

Assumindo que o meio de coordenação é externo e o protocolo é robusto, resulta-se que os forks são
uma preocupação menor do que os ataques de censura.

Além de forks e censura, que exigem ⅓+ poder de votação Bizantina, uma coalizão de +⅔ poder de
voto pode ser pratica arbitrária, estado inválido. Esta é a característica de qualquer sistema
de consenso (BFT). Ao contrário da dupla assinatura, que cria forks com provas facilmente
verificáveis, a detecção de obrigatoriedade de um estado inválido requer que os pares não
validadores verifiquem blocos inteiros, o que implica que eles mantêm uma cópia local do estado
e executam cada transação, computando a raiz de estado de forma independente para eles mesmos.
Uma vez detectado, a única maneira de lidar com essa falha é através do consenso social.
Por exemplo, em situações em que o Bitcoin falhou, seja por causa de bugs de software
(como em março de 2013), ou praticar um estado inválido devido ao comportamento Bizantino
dos mineradores (como em julho de 2015), a comunidade bem conectada de negócios, desenvolvedores,
mineradores e outras organizações estabeleceu um consenso social sobre quais ações manuais se
faziam necessárias para curar a rede. Além disso, uma vez que se pode esperar que os validadores
de uma cadeia de blocos de Tendermint sejam identificáveis, o compromisso de um estado inválido
pode até ser punido por lei ou por alguma jurisprudência externa, se desejado.

### Especificação TMSP

TMSP consiste em 3 tipos de mensagens primárias que são entregues do núcleo para o aplicativo.
O aplicativo responde com mensagens de resposta correspondentes.

A mensagem `AppendTx` é o cavalo de trabalho da aplicação. Cada transação na blockchain
é entregue com esta mensagem. O aplicativo precisa validar cada transação recebida com a
mensagem AppendTx contra o estado atual, o protocolo de aplicativo e as credenciais
criptográficas da transação. Uma transação validada precisa atualizar o estado do
aplicativo - vinculando um valor a um armazenamento de valores chave ou atualizando o banco de dados UTXO.

A mensagem `CheckTx` é semelhante à AppendTx, mas é apenas para validar transações. O mempool do
Tendermint Core primeiro verifica a validade de uma transação com o CheckTx e apenas relata
transações válidas para seus pares. Os aplicativos podem verificar um nonce incremental na transação
e retornar um erro em CheckTx se o nonce é antigo.

A mensagem `Commit` é usada para calcular uma obrigação criptográfica com o estado atual da aplicação,
para ser colocada no próximo cabeçalho do bloco. Isso tem algumas propriedades úteis. Inconsistências
na atualização desse estado agora aparecerão como forks do blockchain que captura uma classe inteira
de erros de programação. Isso também simplifica o desenvolvimento de clientes leves e seguros,
já que as provas de Merkle-hash podem ser provadas verificando o hash de blocos,
e o hash de blocos é assinado por um quórum de validadores (por poder de voto).

Mensagens TMSP adicionais permitem que o aplicativo acompanhe e altere o conjunto
de validadores e que o aplicativo receba as informações do bloco, como a altura e os votos de confirmação.

Pedidos/respostas TMSP são simples mensagens Protobuf.
Confira o [arquivo do esquema](https://github.com/tendermint/abci/blob/master/types/types.proto).

##### AppendTx

- **Arguments**:
  - `Data ([]byte)`: Os bytes de transação solicitada
- **Returns**:
  - `Code (uint32)`: Código de resposta
  - `Data ([]byte)`: Bytes de resultado, se houver
  - `Log (string)`: Debug ou mensagem de erro
- **Usage**:<br/>
  Acrescentar e executar uma transação. Se a transação for válida,
  CodeType.OK

##### CheckTx

- **Arguments**:
  - `Data ([]byte)`: Os bytes de transação solicitados
- **Returns**:
  - `Code (uint32)`: Código de resposta
  - `Data ([]byte)`: Bytes de resultado, se houver
  - `Log (string)`: Debug ou mensagem de erro
- **Usage**:<br/>
  Validar uma transação. Esta mensagem não deve mutar o estado.
  As transações são primeiro executadas através do CheckTx antes da transmissão para os pares na camada mempool.
  Você pode fazer o CheckTx semi-stateful e limpar o estado após `Commit` ou
  `BeginBlock`,
  para permitir sequências dependentes de transações no mesmo bloco.

##### Commit

- **Returns**:
  - `Data ([]byte)`: O hash Merkle raiz
  - `Log (string)`: Debug ou erro de mensagem
- **Usage**:<br/>
  Retorna um hash Merkle raiz do estado da aplicação.

##### Query

- **Arguments**:
  - `Data ([]byte)`: Os bytes de solicitação consultada
- **Returns**:
  - `Code (uint32)`: Código de resposta
  - `Data ([]byte)`: Os bytes de resposta consultada
  - `Log (string)`: Debug ou erro de mensagem

##### Flush

- **Usage**:<br/>
  Limpar a fila de resposta. Aplicações que implementam `types.Application`
  não precisa implementar esta mensagem - é tratada pelo projeto.

##### Info

- **Returns**:
  - `Data ([]byte)`: Os bytes de informação
- **Usage**:<br/>
  Retorna informações sobre o estado da aplicação. Aplicação específicão.

##### SetOption

- **Arguments**:
  - `Key (string)`: Chave para definir
  - `Value (string)`: Valor a definir para a chave
- **Returns**:
  - `Log (string)`: Debug ou mensagem de erro
- **Usage**:<br/>
  Define as opções do aplicativo. Exemplo Key="mode", Value="mempool" para uma conexão mempool
  , ou Key="mode", Value="consensus" para uma conexão de consenso.
  Outras opções são específicas da aplicação.

##### InitChain

- **Arguments**:
  - `Validators ([]Validator)`: validadores de genesis iniciais
- **Usage**:<br/>
  Chamado uma vez na genesis

##### BeginBlock

- **Arguments**:
  - `Height (uint64)`: A altura do bloco que está começando
- **Usage**:<br/>
  Sinaliza o início de um novo bloco. Chamado antes de qualquer AppendTxs.

##### EndBlock

- **Arguments**:
  - `Height (uint64)`: A altura do bloco que terminou
- **Returns**:
  - `Validators ([]Validator)`: Mudança de validadores com novos poderes de voto (0
    para remover)
- **Usage**:<br/>
  Sinaliza o fim de um bloco. Chamado antes de cada Commit após todas as
  transações

Veja [o repositório TMSP](https://github.com/tendermint/abci) para mais detalhes.

### Reconhecimento de entrega de pacotes IBC

Há várias razões pelas quais um remetente pode querer o reconhecimento da entrega de um pacote
pela cadeia de recebimento. Por exemplo, o remetente pode não saber o status da cadeia de
destino, se for esperado que esteja com defeito. Ou, o remetente pode querer impor um tempo
limite no pacote (com o campo `MaxHeight`), enquanto qualquer cadeia de destino pode sofrer
de um ataque de negação de serviço com um aumento repentino no número de pacotes de entrada.

Nesses casos, o remetente pode exigir confirmação de entrega configurando o status
do pacote inicial como `AckPending`. Em seguida, é a responsabilidade da
cadeia receptora confirmar a entrega, incluindo uma abreviada `IBCPacket` no app Merkle hash.

![Figura da Zone1, Zone2, e Hub IBC com
reconhecimento](https://raw.githubusercontent.com/gnuclear/atom-whitepaper/master/msc/ibc_with_ack.png)

Primeiro, um `IBCBlockCommit` e`IBCPacketTx` são postados no "Hub" que prova
a existência de um `IBCPacket` na "Zone1". Digamos que `IBCPacketTx` tem o seguinte valor:

- `FromChainID`: "Zone1"
- `FromBlockHeight`: 100 (say)
- `Packet`: an `IBCPacket`:
  - `Header`: an `IBCPacketHeader`:
    - `SrcChainID`: "Zone1"
    - `DstChainID`: "Zone2"
    - `Number`: 200 (say)
    - `Status`: `AckPending`
    - `Type`: "moeda"
    - `MaxHeight`: 350 (Dizer que "Hub" está atualmente na altura 300)
  - `Payload`: &lt;Os bytes de uma carga paga de "moeda">

Em seguida, um `IBCBlockCommit` e `IBCPacketTx` são publicados na "Zone2" que comprova
a existência de um `IBCPacket` em "Hub". Digamos que `IBCPacketTx` tem o seguinte valor:

- `FromChainID`: "Hub"
- `FromBlockHeight`: 300
- `Packet`: an `IBCPacket`:
  - `Header`: an `IBCPacketHeader`:
    - `SrcChainID`: "Zone1"
    - `DstChainID`: "Zone2"
    - `Number`: 200
    - `Status`: `AckPending`
    - `Type`: "moeda"
    - `MaxHeight`: 350
  - `Payload`: &lt;Os mesmos bytes de uma carga paga de "moeda">

Em seguida, "Zone2" deve incluir em seu app-hash um pacote abreviado que mostra o novo
status de `AckSent`. Um `IBCBlockCommit` e `IBCPacketTx` são colocados de volta no "Hub"
que comprova a existência de um `IBCPacket` abreviado na "Zone2". Digamos que `IBCPacketTx` tem o seguinte valor:

- `FromChainID`: "Zone2"
- `FromBlockHeight`: 400 (say)
- `Packet`: an `IBCPacket`:
  - `Header`: an `IBCPacketHeader`:
    - `SrcChainID`: "Zone1"
    - `DstChainID`: "Zone2"
    - `Number`: 200
    - `Status`: `AckSent`
    - `Type`: "moeda"
    - `MaxHeight`: 350
  - `PayloadHash`: &lt;Os bytes de hash da mesma carga paga de "moeda">

Finalmente, "Hub" deve atualizar o status do pacote de `AckPending` para`AckReceived`.
A evidência desse novo status finalizado deve voltar a "Zone2". Digamos que `IBCPacketTx` tem o seguinte valor:

- `FromChainID`: "Hub"
- `FromBlockHeight`: 301
- `Packet`: an `IBCPacket`:
  - `Header`: an `IBCPacketHeader`:
    - `SrcChainID`: "Zone1"
    - `DstChainID`: "Zone2"
    - `Number`: 200
    - `Status`: `AckReceived`
    - `Type`: "moeda"
    - `MaxHeight`: 350
  - `PayloadHash`: &lt;Os bytes de hash da mesma carga paga de "moeda">

Enquanto isso, "Zone1" pode assumir de maneira otimista a entrega bem-sucedida de um pacote
de "moeda", a menos que provas em contrário sejam comprovadas em "Hub". No exemplo acima,
se "Hub" não tivesse recebido um status `AckSent` de "Zone2" pelo bloco 350, ele teria
definido o status automaticamente para `Timeout`. Essa evidência de um tempo limite pode
ser postada novamente na "Zone1", e quaisquer tokens podem ser retornados.

![Figura da Zone1, Zone2, e Hub IBC com reconhecimento e 
timeout](https://raw.githubusercontent.com/gnuclear/atom-whitepaper/master/msc/ibc_with_ack_timeout.png)

### Árvore Merkle e Especificação de Prova

Existem dois tipos de árvores Merkle suportadas no ecossistema Tendermint / Cosmos: A Árvore Simples e a Árvore IAVL+.

#### Árvore Simples

A Árvore Simples é uma árvore Merkle para uma lista estática de elementos. Se o número de
itens não for um poder de dois, algumas folhas estarão em níveis diferentes. Árvore Simples
tenta manter ambos os lados da árvore da mesma altura, mas a esquerda pode ter um maior.
Esta árvore Merkle é usada para Merkle-lizar as transações de um bloco, e os elementos de
nível superior da raiz do estado do aplicativo.

                    *
                   / \
                 /     \
               /         \
             /             \
            *               *
           / \             / \
          /   \           /   \
         /     \         /     \
        *       *       *       h6
       / \     / \     / \
      h0  h1  h2  h3  h4  h5

      Uma ÁrvoreSimples com sete elementos

#### Árvore IAVL+

O objetivo da estrutura de dados IAVL+ é fornecer armazenamento persistente para pares de valores-chave
no estado do aplicativo, de modo que um hash determinista de raiz Merkle possa ser calculado
eficientemente. A árvore é balanceada usando uma variante do [algoritmo AVL](https://en.wikipedia.org/wiki/AVL_tree), e todas as operações são O(log(n)).

Em uma árvore AVL, as alturas das duas subárvores filhas de qualquer nó diferem por no máximo um.
Sempre que esta condição for violada após uma atualização, a árvore é rebalanceada criando O(log(n))
novos nós que apontam para nós não modificados da árvore antiga. No algoritmo AVL original, os nós
internos também podem conter pares de valores-chave. O algoritmo AVL + (observe o sinal de adição)
modifica o algoritmo AVL para manter todos os valores em folha de nós, enquanto
usando apenas nós de ramo para armazenar chaves. Isso simplifica o algoritmo, mantendo a trilha hash merkle curta.

A Árvore AVL + é análoga à Ethereum [Patricia tries](https://en.wikipedia.org/wiki/Radix_tree).
Há compensações. Chaves não precisam ser hasheadas antes da inserção em árvores IAVL+, portanto,
isso fornece iteração mais rápida ordenada no espaço-chave que pode beneficiar algumas aplicações.
A lógica é mais simples de implementar, requerendo apenas dois tipos de nós - nós internos e nós de folhas.
A prova de Merkle é em média mais curta, sendo uma árvore binária equilibrada. Por outro lado,
a raiz Merkle de uma árvore IAVL+ depende da ordem das atualizações.

Iremos apoiar outras árvores Merkle eficientes, como Patricia Trie, da Ethereum, quando a variante binária estiver disponível.

### Tipos de Transação

Na implementação canônica, as transações são transmitidas para o aplicativo Cosmos hub através da interface TMSP.

O Cosmos Hub aceitará uma série de tipos de transações primárias, incluindo `SendTx`,
`BondTx`, `UnbondTx`, `ReportHackTx`, `SlashTx`, `BurnAtomTx`, `ProposalCreateTx` e `ProposalVoteTx`,
que são relativamente auto-explicativas e será documentado em uma futura revisão deste artigo.
Aqui documentamos os dois principais tipos de transação para IBC: `IBCBlockCommitTx` e `IBCPacketTx`.

#### IBCBlockCommitTx

Uma transação `IBCBlockCommitTx` é composta de:

- `ChainID (string)`: O ID da blockchain
- `BlockHash ([]byte)`: Os bytes de hash de bloco, a raiz Merkle que inclui o app-hash
- `BlockPartsHeader (PartSetHeader)`: Os bytes de cabeçalho do conjunto de blocos,
  apenas necessários para verificar assinaturas de voto
- `BlockHeight (int)`: A altura do commit
- `BlockRound (int)`: A rodada do commit
- `Commit ([]Vote)`: O +⅔ Tendermint `Precommit` de votos que compõem um bloco
- `ValidatorsHash ([]byte)`: O hash da raiz da árvore-Merkle do novo conjunto de validadores
- `ValidatorsHashProof (SimpleProof)`: Uma ÁrvoreSimples da prova-Merkle para provar o
  `ValidatorsHash` contra o `BlockHash`
- `AppHash ([]byte)`: Um hash da raiz da árvore-Merkle da Árvore IAVL do estado de aplicação
- `AppHashProof (SimpleProof)`: Uma ÁrvoreSimples da prova-Merkle para provar o
  `AppHash` contra o `BlockHash`

#### IBCPacketTx

Um `IBCPacket` é composto de:

- `Header (IBCPacketHeader)`: O cabeçalho do pacote
- `Payload ([]byte)`: Os bytes da carga paga do pacote. _Optional_
- `PayloadHash ([]byte)`: O hash para os bytes do pacote. _Optional_

Qualquer um dos `Payload` ou `PayloadHash` deve estar presente. O hash de um `IBCPacket`
é uma raiz Merkle simples dos dois itens, `Header` e `Payload`. Um `IBCPacket` sem a carga completa
é chamado de _abbreviated packet_.

Um `IBCPacketHeader` é composto de:

- `SrcChainID (string)`: O ID da blockchain fonte
- `DstChainID (string)`: O ID da blockchain destino
- `Number (int)`: Um número exclusivo para todos os pacotes
- `Status (enum)`: Pode ser um `AckPending`, `AckSent`, `AckReceived`,
  `NoAck`, ou `Timeout`
- `Type (string)`: Os tipos são dependentes da aplicação. Cosmos reserva-se ao tipo de pacote "moeda"
- `MaxHeight (int)`: Se status não for `NoAckWanted` ou `AckReceived` por essa altura, o status se tornará `Timeout`. _Opcional_

Uma transação `IBCPacketTx` é composta de:

- `FromChainID (string)`: O ID da blockchain que está fornecendo este pacote; Não necessariamente a fonte
- `FromBlockHeight (int)`: A altura da blockchain na qual o seguinte pacote é incluído (Merkle-izado) no hash da blockchain de origem
- `Packet (IBCPacket)`: Um pacote de dados, cujo estado pode ser um
  `AckPending`, `AckSent`, `AckReceived`, `NoAck`, ou `Timeout`
- `PacketProof (IAVLProof)`: Uma prova-Merkle da Árvore IAVL para para provar o hash do pacote contra o \`AppHash' da cadeia de origem em determinada altura

A seqüência para enviar um pacote da "Zone1" para a "Zone2" através do "Hub" é mostrada em {Figure X}.
Primeiro, um `IBCPacketTx` prova ao "Hub" que o pacote está incluído no estado da aplicação de "Zone1".
Em seguida, outro `IBCPacketTx` prova a "Zone2" que o pacote está incluído no estado da aplicação "Hub".
Durante esse procedimento, os campos `IBCPacket` são idênticos: o `SrcChainID` é sempre "Zone1",
e o `DstChainID` é sempre" Zone2 ".

O `PacketProof` deve ter o caminho correto da prova-Merkle, da seguinte maneira:

    IBC/<SrcChainID>/<DstChainID>/<Number>

Quando "Zone1" quer enviar um pacote para "Zone2" através do "Hub", os dados de `IBCPacket`
são idênticos se o pacote é Merkle-izado em "Zone1", no "Hub" ou "Zone2". O único campo mutável
é `Status` para acompanhar a entrega, conforme mostrado abaixo.

## Agradecimentos

Agradecemos aos nossos amigos e colegas por sua ajuda na conceituação, revisão e apoio no nosso trabalho com Tendermint e Cosmos.

- [Zaki Manian](https://github.com/zmanian) da
  [SkuChain](http://www.skuchain.com/) forneceu muita ajuda na formatação e redacção, especialmente sob a seção TMSP
- [Jehan Tremback](https://github.com/jtremback) da Althea and Dustin Byington
  por ajudar com iterações iniciais
- [Andrew Miller](https://soc1024.com/) da [Honey
  Badger](https://eprint.iacr.org/2016/199) pelo feedback sobre consenso
- [Greg Slepak](https://fixingtao.com/) pelo feedback sobre consenso e redação
- Também agradecemos ao [Bill Gleim](https://github.com/gleim) e [Seunghwan
  Han](http://www.seunghwanhan.com) por várias contribuições.
- [Pedro Augusto](https://github.com/ShooterXD) pela tradução para
  Português

## Citações

- [1] Bitcoin: <https://bitcoin.org/bitcoin.pdf>
- [2] ZeroCash: <http://zerocash-project.org/paper>
- [3] Ethereum: <https://github.com/ethereum/wiki/wiki/White-Paper>
- [4] TheDAO: <https://download.slock.it/public/DAO/WhitePaper.pdf>
- [5] Segregated Witness: <https://github.com/bitcoin/bips/blob/master/bip-0141.mediawiki>
- [6] BitcoinNG: <https://arxiv.org/pdf/1510.02037v2.pdf>
- [7] Lightning Network: <https://lightning.network/lightning-network-paper-DRAFT-0.5.pdf>
- [8] Tendermint: <https://github.com/tendermint/tendermint/wiki>
- [9] FLP Impossibility: <https://groups.csail.mit.edu/tds/papers/Lynch/jacm85.pdf>
- [10] Slasher: <https://blog.ethereum.org/2014/01/15/slasher-a-punitive-proof-of-stake-algorithm/>
- [11] PBFT: <http://pmg.csail.mit.edu/papers/osdi99.pdf>
- [12] BitShares: <https://bitshares.org/technology/delegated-proof-of-stake-consensus/>
- [13] Stellar: <https://www.stellar.org/papers/stellar-consensus-protocol.pdf>
- [14] Interledger: <https://interledger.org/rfcs/0001-interledger-architecture/>
- [15] Sidechains: <https://blockstream.com/sidechains.pdf>
- [16] Casper: <https://blog.ethereum.org/2015/08/01/introducing-casper-friendly-ghost/>
- [17] TMSP: <https://github.com/tendermint/abci>
- [18] Ethereum Sharding: <https://github.com/ethereum/EIPs/issues/53>
- [19] LibSwift: <http://www.ds.ewi.tudelft.nl/fileadmin/pds/papers/PerformanceAnalysisOfLibswift.pdf>
- [20] DLS: <http://groups.csail.mit.edu/tds/papers/Lynch/jacm88.pdf>
- [21] Thin Client Security: <https://en.bitcoin.it/wiki/Thin_Client_Security>
- [22] Ethereum 2.0 Mauve Paper: <https://cdn.hackaday.io/files/10879465447136/Mauve%20Paper%20Vitalik.pdf>

#### Links não classificados

- <https://www.docdroid.net/ec7xGzs/314477721-ethereum-platform-review-opportunities-and-challenges-for-private-and-consortium-blockchains.pdf>
