# Cosmos Hub para iniciar Mainnet


En el verano de 2016, se publicó el [Cosmos whitepaper](https://cosmos.network/resources/whitepaper). En la 
primavera de 2017, se completó la [recaudación de fondos de Cosmos](https://fundraiser.cosmos.network/). En los primeros 
meses del 2019, la [última version](https://github.com/cosmos/cosmos-sdk/releases) del _software_  contiene todas las funcionalidades requeridas en la red principal. El lanzamiento del blockchain principal del ecosistema,  el Cosmos Hub, se acerca. 
¿Qué significa esto para los poseedores de Atoms?

Si eres poseedor de Atoms, podrás delegarlos a los validadores de la red principal y votar sobre las propuestas de 
gobernanza. De hecho, el éxito futuro de la red depende de que lo haga de forma responsable. Sin embargo, aún no podrás 
transferir Atoms. Las transferencias se desactivarán a nivel de protocolo hasta que se ejecute un hard-fork para habilitarlas.

Quienes posean Atoms deben seguir cuidadosamente las instrucciones para poder delegar de forma segura. Por favor, 
lee primero toda la guía para familiarizarse con la [línea de comandos](https://github.com/cosmos/cosmos-sdk/blob/develop/docs/gaia/delegator-guide-cli.md)

El proceso descrito en la guía es actualmente la única forma segura y auditada de delegar Atoms en el lanzamiento. Esto se debe a que 
la herramienta gaiacli utilizada en la guía es el único software de cartera que está siendo sometido a auditorías de seguridad 
por parte de terceros en este momento. Ningún otro proveedor de carteras ha iniciado aún las auditorías de seguridad.

Recuerde que delegar Atoms implica un riesgo significativo. Una vez delegados a un validador, los Atoms son bloqueados por un 
período de tiempo durante el cual no pueden ser recuperados. Si el validador se realiza alguna infracción durante este tiempo, algunos o 
todos los Atoms delegados pueden ser destruidos. Es su responsabilidad realizar la debida diligencia con los validadores antes 
de delegar!

The Cosmos Hub es un software altamente experimental. En estos primeros días, podemos esperar tener problemas, actualizaciones 
y errores. Las herramientas existentes requieren conocimientos técnicos avanzados e implican riesgos que están fuera del control 
de la Fundación Interchain y/o del equipo de Tendermint (ver también la sección de riesgos en [Interchain Cosmos Contribution Terms](https://github.com/cosmos/cosmos/blob/master/fundraiser/Interchain%20Cosmos%20Contribution%20Terms%20-%20FINAL.pdf)). 
Cualquier uso de este software de código abierto [con licencia Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0) se hace bajo su propio riesgo y sobre una 
base "TAL CUAL" sin garantías ni condiciones de ningún tipo, y toda responsabilidad de la Fundación Interchain y/o del equipo 
de Tendermint por daños y perjuicios que surjan en relación con el software está excluida. ¡Por favor procede con precaución!

Si buscas más información sobre la delegación y quieres hablar con la gente que está desarrollando Cosmos, únete a la reunión 
virtual del 14 de febrero, en la que se explicarán paso a paso las instrucciones para delegar Atoms en el lanzamiento.

Regístrese aquí: [gotowebinar.com/register/](https://register.gotowebinar.com/register/5028753165739687691)

## Hitos restantes para el lanzamiento

Sigue en el [sitio web oficial](https://cosmos.network/launch), para seguir el progreso del lanzamiento de la red principal.

### 5 Cosmos-SDK Auditorías de seguridad ✔

A principios de enero, la Cosmos-SDK se sometió a la primera de una serie de evaluaciones de seguridad de terceros programadas para el primer trimestre de 2019. Esta auditoría se llevó a cabo durante un período de dos semanas y media. Hasta la fecha, 
dos empresas de auditoría de seguridad diferentes han evaluado varias partes de la SDK y actualmente se está llevando a cabo una 
tercera auditoría.

### 4 Cosmos SDK Congelamiento de funcionalidades extra

Los últimos cambios importantes del Cosmos-SDK están incluidos en el [v0.31.0 launch RC](https://github.com/cosmos/cosmos-sdk/projects/27). 
Una vez que se complete este candidato de lanzamiento, el equipo de la Cosmos-SDK realizará internamente  
una diligencia debida de seguridad suficiente previa al lanzamiento.

Inmediatamente después del lanzamiento de la version 0.31.0 de la Cosmos-SDK, se lanzará una _testnet_ de Gaia con el fin de
eliminar cualquier _bug_.

### 3 Game of Stakes Completado

Game of Stakes (GoS),[el primer concurso de redes de prueba adversarias de su tipo](https://blog.cosmos.network/announcing-incentivized-testnet-game-efe64e0956f6), 
se lanzó en diciembre de 2018 para poner a prueba los incentivos económicos y los estratos sociales de un blockchain asegurado exclusivamente
por Proof-of-Stake. Hasta la fecha, la cadena de bloques de GoS ha tenido _hard-forks_ con éxito en tres ocasiones.
Tan pronto cuando GoS concluya, se utilizarán los [criterios de puntuación](https://github.com/cosmos/game-of-stakes/blob/master/README.md#scoring) para determinar a
los ganadores. Estos serán anunciados después de la finalización del juego.

### 2 Transacciones para el Genesis recopiladas

La Fundación Interchain publicará una recomendación para la asignación de Atoms en el génesis. Esto incluirá asignaciones 
para los participantes de la recaudación de fondos de Cosmos, los primeros contribuyentes y los ganadores del Game of Stakes. 
Cualquier persona con una asignación recomendada tendrá la oportunidad de presentar una _gentx_ (transacción génesis), el cual es requerida para 
convertirse en validador en el génesis de la red. El resultado final de la asignación recomendada y la colección de _gentxs_ es 
un [archivo de génesis](https://forum.cosmos.network/t/genesis-files-network-starts-vs-upgrades/1464) final.

### 1 Lanzamiento del Cosmos Hub en Mainnet

Una vez que el archivo de génesis es aprobado por la comunidad, y +⅔ del poder de voto se pone en línea, la red principal 
de Cosmos habrá oficialmente realizado su lanzamiento.

## Canales Oficiales de comunicacion de Cosmos

Estas son las cuentas oficiales por las que se comunicarán los detalles del lanzamiento:

- [Cosmos Network](https://twitter.com/cosmos)
- [Cosmos GitHub](https://github.com/cosmos)
- [Cosmos Blog](https://blog.cosmos.network)

Por favor, tenga en cuenta que el [foro](https://forum.cosmos.network/),[grupos de chat de RIOT](https://riot.im/app/#/group/+cosmos:matrix.org), y [grupo de Telegram](http://t.me/cosmosproject)
no deben ser tratados como noticias oficiales de Cosmos.

Si tiene dudas o confusión sobre los próximos pasos a seguir y no está seguro sobre fuentes de información confiables, 
no haga nada durante el período inicial y espere una actualización a través de los tres canales de comunicación mencionados 
anteriormente. No proporciones nunca tus 12 palabras secretas a ningún administrador, sitio web o _software_ no oficial.

**Nunca le pediremos su clave privada o sus palabras de recuperación.**

## Manténgase seguro (y protegido) para el lanzamiento de la Mainnet

El lanzamiento de cualquier _blockchain_ público es un momento increíblemente emocionante, y definitivamente es uno que 
los actores maliciosos pueden tratar de aprovechar para su propio beneficio personal. La [ingeniería social](https://en.wikipedia.org/wiki/Social_engineering_%28security%29) ha 
existido aproximadamente desde que los seres humanos han estado en el planeta, y en la era técnica, suele adoptar la forma 
de [phishing](https://ssd.eff.org/en/module/how-avoid-phishing-attacks) o [spearphishing](https://en.wikipedia.org/wiki/Phishing#Spear_phishing). Ambos ataques son formas de engaño muy exitosas que son responsables de más del 95% de las 
infracciones de seguridad de las cuentas, y no sólo se producen a través del correo electrónico: en la actualidad, se 
producen intentos de phishing oportunistas y selectivos [en cualquier lugar en el que haya una bandeja de entrada](https://www.umass.edu/it/security/phishing-fraudulent-emails-text-messages-phone-calls). 
No importa si está usando Signal, Telegram, SMS, Twitter, o simplemente revisando sus DMs en foros o redes sociales, los 
atacantes tienen una [plétora de oportunidades](https://jia.sipa.columbia.edu/weaponization-social-media-spear-phishing-and-cyberattacks-democracy) para entrar en su vida digital en un esfuerzo por separarlo de la 
información valiosa y de los activos que tiene en su poder. definitivamente no quieren perder.

Aunque la perspectiva de tener que tratar con un actor malicioso que conspira contra ti puede parecer desalentadora, hay 
muchas cosas que puedes hacer para protegerte de todo tipo de esquemas de ingeniería social. En cuanto a la preparación para 
el lanzamiento de Mainnet, esto debería requerir entrenar tus instintos para tener éxito. detectar y evitar los riesgos de 
seguridad, seleccionando recursos que sirvan como fuente de verdad para verificar la información, y pasar por algunos pasos 
técnicos para  reducir o eliminar el riesgo de robo de claves o credenciales.

**Estas son algunas de las reglas de participación que debes tener en cuenta cuando te prepares para el lanzamiento de Cosmos Mainnet:**

- Descarga software directamente de fuentes oficiales, y asegúrate de que siempre estás usando la última y más segura versión 
de gaiacli cuando estés haciendo algo que incluya tus 12 palabras. Las últimas versiones de Tendermint, el Cosmos-SDK y gaiacli 
siempre estarán disponibles en nuestros repositorios oficiales de GitHub, y descargándolas desde allí se garantiza que no te 
engañarán para que utilices una versión maliciosamente modificada del software.

- No comparta sus 12 palabras con nadie. La única persona que debería conocerlos es usted. Esto es especialmente importante 
si alguna vez alguien se le acerca para ofrecerle servicios de custodia para sus Atoms: para evitar perder el control de sus 
tokens, debe almacenarlos fuera de línea para minimizar el riesgo de robo y tener una fuerte estrategia de copia de seguridad 
en marcha. Y nunca, nunca, nunca los comparta con nadie más.

- Sea escéptico ante adjuntos inesperados o correos electrónicos que le pidan que visite un sitio web sospechoso o 
desconocido en el contexto de cadenas de bloques o cryptocurrency. Un atacante puede intentar atraerlo a 
un [sitio comprometido](https://blog.malwarebytes.com/cybercrime/2013/02/tools-of-the-trade-exploit-kits/) diseñado para robar información confidencial de su equipo. Si eres usuario de Gmail, prueba tu 
capacidad de recuperación frente a las últimas tácticas de phishing basadas en correo electrónico[aquí](https://phishingquiz.withgoogle.com/).

- Haga su debida diligencia antes de comprar Atoms. Los Atoms no serán transferibles en el momento del lanzamiento, por lo 
que no pueden ser comprados o vendidos hasta que un hard-fork lo permita. Cuando sean transferibles, asegúrese de que ha 
investigado al vendedor o intercambiador para confirmar que los Atoms provienen de una fuente confiable.

- Ni el equipo de Tendermint ni la Fundación Interchain venderán Atoms, así que si ves mensajes en los medios sociales o 
correos electrónicos que anuncian una venta simbólica de nosotros, no son reales y deben ser evitados.  Habilite la 
autenticación de 2 factores y tenga en cuenta los métodos de recuperación utilizados para recuperar el acceso a sus cuentas 
más importantes. Las cuentas sin protección como el correo electrónico, los medios sociales, tu cuenta de GitHub, el Foro 
Cosmos y cualquier otra cosa en el medio podrían dar a un atacante oportunidades de ganar terreno en tu vida en línea. Si 
aún no lo ha hecho, empiece a utilizar inmediatamente una aplicación de autenticación o una hardware key dondequiera que 
manejes tus tokens.  Esta es una manera simple, efectiva y comprobada para reducir el riesgo de robo de cuenta.

- Sea escéptico respecto a los consejos técnicos, especialmente si provienen de personas que no conoce en los foros y en los 
canales de chat grupal. Familiarícese con los comandos importantes, especialmente aquellos que le ayudarán a llevar a cabo 
acciones de alto riesgo, y consulte nuestra documentación oficial para asegurarse de que no se le está engañando para que 
haga algo que pueda perjudicarle a usted o a su validador. Y recuerda que el foro del Cosmos, los canales Riot y Telegram no 
son fuentes de información oficial o de noticias sobre Cosmos. 

- Verifique las transacciones antes de pulsar enviar. Sí, esas secuencias de direcciones son largas, pero comparándolas 
visualmente en bloques de 4 carácteres a la vez puede ser la diferencia entre enviarlas al lugar correcto o enviarlas al 
olvido.

*Si aparece un acuerdo que [suena demasiado bueno para ser cierto][https://www.psychologytoday.com/us/blog/mind-in-the-machine/201712/how-fear-is-being-used-manipulate-cryptocurrency-markets], o aparece un mensaje pidiéndote información que 
nunca, nunca, nunca debe ser compartida con otra persona, siempre puedes verificarla antes de comprometerte 
con ella al visitar un sitio web o canal de comunicación oficial de Cosmos por su cuenta. Nadie de Cosmos, el equipo de 
Tendermint o la Fundación Interchain enviarán alguna vez un email pidiéndole sus 12 palabras secretas para compartirlas con nosotros, 
y siempre usaremos nuestro blog oficial, Twitter y GitHub para comunicar noticias importantes directamente a la comunidad 
del Cosmos.*

[whitepaper]: https://cosmos.network/resources/whitepaper
[fundraiser]: https://fundraiser.cosmos.network/
[releases]: https://github.com/cosmos/cosmos-sdk/releases
[cosmos]: https://cosmos.network/launch
[social]: https://en.wikipedia.org/wiki/Social_engineering_%28security%29
[phishing]: https://ssd.eff.org/en/module/how-avoid-phishing-attacks
[spearphishing]: https://en.wikipedia.org/wiki/Phishing#Spear_phishing
[inbox]: https://www.umass.edu/it/security/phishing-fraudulent-emails-text-messages-phone-calls
[opportunities]: https://jia.sipa.columbia.edu/weaponization-social-media-spear-phishing-and-cyberattacks-democracy
[cli]: https://github.com/cosmos/cosmos-sdk/blob/develop/docs/gaia/delegator-guide-cli.md
[webinar]: https://register.gotowebinar.com/register/5028753165739687691
[terms]: https://github.com/cosmos/cosmos/blob/master/fundraiser/Interchain%20Cosmos%20Contribution%20Terms%20-%20FINAL.pdf
[apache]: https://www.apache.org/licenses/LICENSE-2.0
[gos]: https://blog.cosmos.network/announcing-incentivized-testnet-game-efe64e0956f6
[scoring]: https://github.com/cosmos/game-of-stakes/blob/master/README.md#scoring
[file]: https://forum.cosmos.network/t/genesis-files-network-starts-vs-upgrades/1464
[forum]: https://forum.cosmos.network/
[riot]: https://riot.im/app/#/group/+cosmos:matrix.org
[telegram]: http://t.me/cosmosproject
[good]: https://www.psychologytoday.com/us/blog/mind-in-the-machine/201712/how-fear-is-being-used-manipulate-cryptocurrency-markets
[rc]: https://github.com/cosmos/cosmos-sdk/projects/27
[compromised site]: https://blog.malwarebytes.com/cybercrime/2013/02/tools-of-the-trade-exploit-kits/
[quiz]: https://phishingquiz.withgoogle.com/
