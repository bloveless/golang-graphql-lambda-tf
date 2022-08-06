let SessionLoad = 1
let s:so_save = &g:so | let s:siso_save = &g:siso | setg so=0 siso=0 | setl so=-1 siso=-1
let v:this_session=expand("<sfile>:p")
silent only
silent tabonly
cd ~/Projects/golang-graphql-api-tf
if expand('%') == '' && !&modified && line('$') <= 1 && getline(1) == ''
  let s:wipebuf = bufnr('%')
endif
let s:shortmess_save = &shortmess
if &shortmess =~ 'A'
  set shortmess=aoOA
else
  set shortmess=aoO
endif
badd +17 Makefile
badd +158 infra/main.tf
badd +1 infra/lambda.tf
badd +52 infra/apigateway.tf
badd +39 main.go
badd +12 ~/.asdf/installs/golang/1.19/packages/pkg/mod/github.com/aws/aws-lambda-go@v1.34.1/events/activemq.go
badd +486 infra/terraform.tfstate
badd +1 infra/.terraform.lock.hcl
badd +1 infra/dynamodb.tf
badd +203 ~/.asdf/installs/golang/1.18.3/packages/pkg/mod/github.com/graph-gophers/graphql-go@v1.4.0/graphql.go
badd +18 ~/.asdf/installs/golang/1.18.3/packages/pkg/mod/github.com/graph-gophers/graphql-go@v1.4.0/internal/query/query.go
badd +10 ~/.asdf/installs/golang/1.18.3/packages/pkg/mod/github.com/graph-gophers/graphql-go@v1.4.0/types/query.go
badd +17 ~/.asdf/installs/golang/1.18.3/packages/pkg/mod/github.com/aws/aws-lambda-go@v1.34.1/events/apigw.go
badd +469 ~/.asdf/installs/golang/1.18.3/go/src/encoding/base64/base64.go
badd +52 ~/.asdf/installs/golang/1.18.3/packages/pkg/mod/github.com/aws/aws-sdk-go@v1.44.70/service/dynamodb/service.go
badd +17588 ~/.asdf/installs/golang/1.18.3/packages/pkg/mod/github.com/aws/aws-sdk-go@v1.44.70/service/dynamodb/api.go
badd +9 ~/.asdf/installs/golang/1.18.3/packages/pkg/mod/github.com/graph-gophers/graphql-go@v1.4.0/id.go
argglobal
%argdel
$argadd ~/Projects/golang-graphql-api-tf/
edit infra/main.tf
argglobal
balt main.go
setlocal fdm=manual
setlocal fde=0
setlocal fmr={{{,}}}
setlocal fdi=#
setlocal fdl=0
setlocal fml=1
setlocal fdn=20
setlocal fen
silent! normal! zE
let &fdl = &fdl
let s:l = 1 - ((0 * winheight(0) + 26) / 52)
if s:l < 1 | let s:l = 1 | endif
keepjumps exe s:l
normal! zt
keepjumps 1
normal! 0
lcd ~/Projects/golang-graphql-api-tf
tabnext 1
if exists('s:wipebuf') && len(win_findbuf(s:wipebuf)) == 0 && getbufvar(s:wipebuf, '&buftype') isnot# 'terminal'
  silent exe 'bwipe ' . s:wipebuf
endif
unlet! s:wipebuf
set winheight=1 winwidth=20
let &shortmess = s:shortmess_save
let s:sx = expand("<sfile>:p:r")."x.vim"
if filereadable(s:sx)
  exe "source " . fnameescape(s:sx)
endif
let &g:so = s:so_save | let &g:siso = s:siso_save
set hlsearch
let g:this_session = v:this_session
let g:this_obsession = v:this_session
doautoall SessionLoadPost
unlet SessionLoad
" vim: set ft=vim :
