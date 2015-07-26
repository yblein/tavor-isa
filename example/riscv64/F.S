flw       @f, $i12(@r)
fsw       @f, $i12(@r)
fmadd.s   @f, @f, @f, @f
fmsub.s   @f, @f, @f, @f
fnmsub.s  @f, @f, @f, @f
fnmadd.s  @f, @f, @f, @f
fadd.s    @f, @f, @f
fsub.s    @f, @f, @f
fmul.s    @f, @f, @f
fdiv.s    @f, @f, @f
fsqrt.s   @f, @f
fsgnj.s   @f, @f, @f
fsgnjn.s  @f, @f, @f
fsgnjx.s  @f, @f, @f
fmin.s    @f, @f, @f
fmax.s    @f, @f, @f
fcvt.w.s  @r, @f
fcvt.wu.s @r, @f
fmv.x.s   @r, @f
feq.s     @r, @f, @f
flt.s     @r, @f, @f
fle.s     @r, @f, @f
fclass.s  @r, @f
fcvt.s.w  @f, @r
fcvt.s.wu @f, @r
fmv.s.x   @f, @r
frcsr     @r
frrm      @r
frflags   @r
fscsr     @r, @r
fsrm      @r, @r
fsflags   @r, @r
#fsrmi     @r, $i12
#fsflagsi  @r, $i12
fcvt.l.s  @r, @f
fcvt.lu.s @r, @f
fcvt.s.l  @f, @r
fcvt.s.lu @f, @r
