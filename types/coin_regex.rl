package types

func MatchDenom(data []byte) bool {
%% machine scanner;
%% write data;
    cs, p, pe, eof := 0, 0, len(data), len(data)
    _ = eof
    %%{
        # Define character classes
        special = '/' | ':' | '.' | '_' | '-';

        denom_pattern = [a-zA-Z] (alnum | special){2,127};


        # Combined pattern for matching either a denomination or a decimal amount
        main := denom_pattern  @{ return true };

        write init;
        write exec;
    }%%
    return false
}

func MatchDecCoin(data []byte) (amountStart, amountEnd, denomEnd int, isValid bool) {
    %% machine dec_coin;
    %% write data;

    // Initialize positions and validity flag
    amountStart, amountEnd, denomEnd = -1, -1, -1
    isValid = false

    // Ragel state variables
    var cs, p, pe, eof int
    p, pe, eof = 0, len(data), len(data)

    %%{
        action StartAmount {
            amountStart = p;
        }
        action EndAmount {
            amountEnd = p;
        }
        action EndDenom {
            denomEnd = p-1; // Adjusted to exclude space if present
        }
        action MarkValid {
            isValid = true;
        }

        special = '/' | ':' | '.' | '_' | '-';

        dec_amt = (digit+ >StartAmount ('.' digit+)? %EndAmount) | ('.' >StartAmount digit+ %EndAmount);

        denom = [a-zA-Z] (alnum | special){2,127} %EndDenom;

        main := dec_amt (space* denom)?;

        write init;
        write exec;
    }%%

    isValid = (cs >= %%{ write first_final; }%%);

    // Return the captured positions and validity
    return amountStart, amountEnd, denomEnd, isValid
}
