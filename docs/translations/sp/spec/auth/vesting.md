# Vesting

 - [Vesting](#vesting)
  - [Introducción y requisitos](#Introducción-y-requisitos)
  - [Tipos de vesting account](#Tipos-de-vesting-account)
  - [Especificación del vesting account](#Especificación-del-vesting-account)
    - [Especificación de importes Vesting & Vested](#Especificación-de-importes-Vesting-&-Vested)
      - [Vesting Accounts continuadas](#Vesting-Accounts-continuadas)
      - [Vesting Accounts Retrasadas/Separadas](#Vesting-Accounts-RetrasadasSeparadas)
    - [Transferir/Enviar](#TransferirEnviar)
      - [Guardianes/Responsables](#guardianesresponsables)
    - [Delegando](#delegando)
      - [Guardianes/Responsables](#guardianesresponsables-1)
    - [Undelegating](#undelegating)
      - [Guardianes/Responsables](#guardianesresponsables-2)
  - [Guardianes & Responsables](#guardianes--responsables)
  - [Inicialización del genesis](#Inicialización-del-genesis)
  - [Ejemplos](#ejemplos)
    - [Simple](#simple)
    - [Recortes](#recortes)
  - [Glosario](#glosario)

 ## Introducción y requisitos

 Esta especificación describe la implementación del vesting account para el Cosmos Hub.
Los requisitos para esta cuenta de titularidad son que se deben inicializar durante el génesis con un 
balance inicial `X` y un tiempo final de adquisición de derechos `T`.

 El propietario de esta cuenta debe poder delegar en validadores y no delegar en ellos; sin embargo, 
no puede enviar monedas bloqueadas a otras cuentas hasta que dichas monedas hayan sido totalmente transferidas.

 Además, una vesting account confiere a todas sus denominaciones de moneda al mismo tiempo
tasa de interés. Esto puede estar sujeto a cambios.

 **Nota**: Una vesting account podría tener algunas monedas con y sin derechos adquiridos. Para
el soporte de tal característica, el tipo `GenesisAccount` tendrá que ser actualizado para hacer tal distinción.

 ## Tipos de vesting account:

 ```go
// VestingAccount defines an interface that any vesting account type must
// implement.
type VestingAccount interface {
    Account
     GetVestedCoins(Time)  Coins
    GetVestingCoins(Time) Coins
     // Delegation and undelegation accounting that returns the resulting base
    // coins amount.
    TrackDelegation(Time, Coins)
    TrackUndelegation(Coins)
     GetStartTime() int64
    GetEndTime()   int64
}
 // BaseVestingAccount implements the VestingAccount interface. It contains all
// the necessary fields needed for any vesting account implementation.
type BaseVestingAccount struct {
    BaseAccount
     OriginalVesting  Coins // coins in account upon initialization
    DelegatedFree    Coins // coins that are vested and delegated
    DelegatedVesting Coins // coins that vesting and delegated
     EndTime  int64 // when the coins become unlocked
}
 // ContinuousVestingAccount implements the VestingAccount interface. It
// continuously vests by unlocking coins linearly with respect to time.
type ContinuousVestingAccount struct {
    BaseVestingAccount
     StartTime  int64 // when the coins start to vest
}
 // DelayedVestingAccount implements the VestingAccount interface. It vests all
// coins after a specific time, but non prior. In other words, it keeps them
// locked until a specified time.
type DelayedVestingAccount struct {
    BaseVestingAccount
}
```

 Con el fin de facilitar la comprobación y las afirmaciones de tipo ad-hoc y de apoyar
flexibilidad en el uso de la cuenta, la interfaz `Account` existente se actualiza para contener
lo siguiente:

 ```go
type Account interface {
    // ...
     // Calculates the amount of coins that can be sent to other accounts given
    // the current time.
    SpendableCoins(Time) Coins
}
```

 ## Especificación del vesting account

 Dada una vesting account, definimos lo siguiente en las operaciones en curso:

 - `OV`: El importe original de la cantidad de moneda que se adquiere. Es un valor constante.
- `V`: El número de monedas `OV` que todavía están _vesting_. Se deriva de `OV`, `StartTime` 
y `EndTime`. Este valor se calcula a petición y no por bloque.
- `V'`: El número de monedas `OV` que están _vested_ (desbloqueadas). Este valor se calcula 
a petición y no por bloque.
- `DV`: El número de monedas delegadas _vesting_. Es un valor variable. Se almacena y modifica 
directamente en la vesting account.
- `DF`: El número de monedas delegadas _vested_ (desbloqueadas). Es un valor variable. Se almacena 
y modifica directamente en la cuenta de titularidad.
- `BC`: El número de monedas `OV` menos las monedas transferidas (que pueden ser negativas o delegadas). 
Se considera que es el saldo de la cuenta base incorporada. Se almacena y modifica directamente en la cuenta de titularidad.

 ### Especificación de importes Vesting & Vested

 Es importante tener en cuenta que estos valores se calculan a petición y no obligatoriamente por bloque (e.g. 
`BeginBlocker` o `EndBlocker`).

 #### Vesting Accounts continuadas

 Para determinar la cantidad de monedas que se adquieren para un tiempo de bloque `T`, el parámetro realiza el 
siguiente paso:

 1. Calcula `X := T - StartTime`
2. Calcula `Y := EndTime - StartTime`
3. Calcula `V' := OV * (X / Y)`
4. Calcula `V := OV - V'`

 Por lo tanto, la cantidad total de monedas _vested_ es `V'` y la cantidad restante, `V` es _vesting_.

 ```go
func (cva ContinuousVestingAccount) GetVestedCoins(t Time) Coins {
    if t <= cva.StartTime {
        // We must handle the case where the start time for a vesting account has
        // been set into the future or when the start of the chain is not exactly
        // known.
        return ZeroCoins
    } else if t >= cva.EndTime {
        return cva.OriginalVesting
    }
     x := t - cva.StartTime
    y := cva.EndTime - cva.StartTime
     return cva.OriginalVesting * (x / y)
}
 func (cva ContinuousVestingAccount) GetVestingCoins(t Time) Coins {
    return cva.OriginalVesting - cva.GetVestedCoins(t)
}
```

 #### Vesting Accounts Retrasadas/Separadas

 Delayed vesting accounts son más fáciles de razonar, ya que sólo tienen el número completo de
cantidad vesting hasta cierto momento, entonces todas las monedas se convierten en vested 
(desbloqueados). Esto no incluye las monedas desbloqueadas que la cuenta pueda tener inicialmente.

 ```go
func (dva DelayedVestingAccount) GetVestedCoins(t Time) Coins {
    if t >= dva.EndTime {
        return dva.OriginalVesting
    }
     return ZeroCoins
}
 func (dva DelayedVestingAccount) GetVestingCoins(t Time) Coins {
    return dva.OriginalVesting - dva.GetVestedCoins(t)
}
```

 ### Transferir/Enviar

 En cualquier momento, una vesting account puede transferir: `min((BC + DV) - V, BC)`.

 En otras palabras, una vesting account puede transferir el mínimo de la cuenta base
y el saldo de la cuenta base más el número de cuentas actualmente delegadas
las monedas con derecho a voto menos el número de monedas que se han concedido hasta la fecha.

 ```go
func (va VestingAccount) SpendableCoins(t Time) Coins {
    bc := va.GetCoins()
    return min((bc + va.DelegatedVesting) - va.GetVestingCoins(t), bc)
}
```

 #### Guardianes/Responsables

 El poseedor correspondiente del `x/bank` debe manejar adecuadamente el envío de las monedas.
basado en si es una vesting account o no.

 ```go
func SendCoins(t Time, from Account, to Account, amount Coins) {
    bc := from.GetCoins()
     if isVesting(from) {
        sc := from.SpendableCoins(t)
        assert(amount <= sc)
    }
     newCoins := bc - amount
    assert(newCoins >= 0)
     from.SetCoins(bc - amount)
    to.SetCoins(amount)
     // save accounts...
}
```

 ## Delegando

 En el caso de una vesting account intente delegar monedas `D`, se realiza lo siguiente:

 1. Verifica `BC >= D > 0`
2. Calcula `X := min(max(V - DV, 0), D)` (portion of `D` that is vesting)
3. Calcula `Y := D - X` (portion of `D` that is free)
4. Establece `DV += X`
5. Establece `DF += Y`
6. Establece `BC -= D`

 ```go
func (va VestingAccount) TrackDelegation(t Time, amount Coins) {
    x := min(max(va.GetVestingCoins(t) - va.DelegatedVesting, 0), amount)
    y := amount - x
     va.DelegatedVesting += x
    va.DelegatedFree += y
    va.SetCoins(va.GetCoins() - amount)
}
```

 #### Guardianes/Responsables

 ```go
func DelegateCoins(t Time, from Account, amount Coins) {
    bc := from.GetCoins()
    assert(amount <= bc)
     if isVesting(from) {
        from.TrackDelegation(t, amount)
    } else {
        from.SetCoins(sc - amount)
    }
     // save account...
}
```

 ### Undelegating

 En el caso de una vesting account que intente no delegar las monedas `D`, se realiza lo siguiente:

 1. Verifica `(DV + DF) >= D > 0` (this is simply a sanity check)
2. Calcula `X := min(DF, D)` (portion of `D` that should become free, prioritizing free coins)
3. Calcula `Y := D - X` (portion of `D` that should remain vesting)
4. Establece `DF -= X`
5. Establece `DV -= Y`
6. Establece `BC += D`

 ```go
func (cva ContinuousVestingAccount) TrackUndelegation(amount Coins) {
    x := min(cva.DelegatedFree, amount)
    y := amount - x
     cva.DelegatedFree -= x
    cva.DelegatedVesting -= y
    cva.SetCoins(cva.GetCoins() + amount)
}
```

 **Nota**: Si se recorta una delegación, la vesting account terminará con una cantidad excesiva de `DV`, 
incluso después de que todas sus monedas hayan sido vested. Esto se debe a que se da prioridad a las 
monedas libres no delegadas. 

 #### Keepers/Handlers

 ```go
func UndelegateCoins(to Account, amount Coins) {
    if isVesting(to) {
        if to.DelegatedFree + to.DelegatedVesting >= amount {
            to.TrackUndelegation(amount)
            // save account ...
        }
    } else {
        AddCoins(to, amount)
        // save account...
    }
}
```

 ## Keepers & Handlers

 Las implementaciones de `VestingAccount` residen en `x/auth`. Sin embargo, cualquier guardian en un módulo 
(por ejemplo, una participación en `x/staking`) que desee utilizar cualquier tipo de adquisición de derechos
monedas, debe llamar a los métodos explícitos en el poseedor de las `x/banco` (por ejemplo, `DelegateCoins`)
opuesto a `SendCoins` y `SubtractCoins`.

 Además, la vesting account también debe poder gastar las monedas que reciba de otros usuarios. Por lo tanto, 
el controlador `MsgSend` del módulo bancario debería cometer un error si una cuenta con derechos adquiridos 
está intentando enviar una cantidad que excede la cantidad de la moneda desbloqueada.

 Consulte la especificación anterior para obtener más detalles sobre la implementación.

 ## Genesis Initialization

 Para inicializar tanto las vesting accounts como las non-vesting accounts, la estructura 
`GenesisAccount` permite incluir nuevos campos: `Vesting`,`StartTime`, y `EndTime`.  Las cuentas destinadas a       
ser de tipo `BaseAccount` o cualquier otro tipo de cuentas non-vesting tendrán `Vesting = false`. La lógica de 
inicialización del génesis (por ejemplo, "initFromGenesisState") tendrá que analizar y devolver las cuentas 
correctas basadas en estos nuevos campos.

 ```go
type GenesisAccount struct {
    // ...
     // vesting account fields
    OriginalVesting  sdk.Coins `json:"original_vesting"`
    DelegatedFree    sdk.Coins `json:"delegated_free"`
    DelegatedVesting sdk.Coins `json:"delegated_vesting"`
    StartTime        int64     `json:"start_time"`
    EndTime          int64     `json:"end_time"`
}
 func ToAccount(gacc GenesisAccount) Account {
    bacc := NewBaseAccount(gacc)
     if gacc.OriginalVesting > 0 {
        if ga.StartTime != 0 && ga.EndTime != 0 {
            // return a continuous vesting account
        } else if ga.EndTime != 0 {
            // return a delayed vesting account
        } else {
            // invalid genesis vesting account provided
            panic()
        }
    }
     return bacc
}
```

 ## Ejemplos

 ### Simple 

 Dada una vesting account continua con 10 monedas de consolidación.

 ```
OV = 10
DF = 0
DV = 0
BC = 10
V = 10
V' = 0
```

 1. Inmediatamente recibe 1 moneda
    ```text
    BC = 11
    ```
2. El tiempo pasa, 2 monedas otorgadas
    ```text
    V = 8
    V' = 2
    ```
3. Delega 4 monedas al validador A
    ```text
    DV = 4
    BC = 7
    ```
4. Envía 3 monedas
    ```text
    BC = 4
    ```
5. Pasa ms tiempo, 2 monedas más otorgadas 
    ```text
    V = 6
    V' = 4
    ```
6. Envía 2 monedas. En este punto, la cuenta ya no puede enviar más monedas hasta que no se hayan depositado más 
o hasta que reciba monedas adicionales. Sin embargo todavía puede hacerlo, delegando.
    ```text
    BC = 2
    ```

 ### Recortes

 Las mismas condiciones de inicio que en el ejemplo simple.

 1. El tiempo pasa, 5 monedas delegadas
    ```text
    V = 5
    V' = 5
    ```
2. Delegar 5 monedas al validador A
    ```text
    DV = 5
    BC = 5
    ```
3. Delegar 5 monedas al validador B
    ```text
    DF = 5
    BC = 0
    ```
4. El validador A se reduce en un 50%, lo que hace que la delegación a A ahora valga 2,5 monedas.
5. Desdelegando desde el validador A (2.5 monedas)
    ```text
    DF = 5 - 2.5 = 2.5
    BC = 0 + 2.5 = 2.5
    ```
6. Desdelegando desde el validador B (5 monedas). La cuenta en este punto sólo puede enviar 
2,5 monedas a menos que reciba más monedas o hasta que más monedas sean entregadas. Sin embargo, todavía 
puede, delegando. 
    ```text
    DV = 5 - 2.5 = 2.5
    DF = 2.5 - 2.5 = 0
    BC = 2.5 + 5 = 7.5
    ```

 Observe que tenemos una cantidad excesiva de `DV`

 ## Glosario

 - OriginalVesting: La cantidad de monedas (por designación) que forman parte inicialmente de una vesting account. 
Estas monedas están fijadas en genesis.
- StartTime: El momento del BFT en el cual una vesting account comienza a percibir derechos.
- EndTime: El momento del BFT en el que una vesting account está totalmente establecida.
- DelegatedFree: La cantidad de monedas (por designación) que se delegan de una vesting account y 
que han sido totalmente otorgados en el momento de la delegación.
- DelegatedVesting: La cantidad de monedas rastreadas (por designación) que se delegan de una vesting account 
otorgados en el momento de la delegación.
- ContinuousVestingAccount: Una implementación de la vesting account que confiere monedas linealmente a lo largo del tiempo.
