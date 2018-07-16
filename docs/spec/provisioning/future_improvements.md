## Future Improvements

### Power Change

Within the current implementation all power changes are held within the
The set of all `powerChange`
may be trimmed from its oldest members once all validators have synced past the
height of the oldest `powerChange`.  This trim procedure will occur on an epoch
basis.  

