# Turns relay.ico into the Windows resource files (rsrc_windows_*.syso) that put the icon on the .exe.
import struct, sys, os
OUT=os.path.join(os.path.dirname(os.path.abspath(__file__)),'..')
ico=open(os.path.join(OUT,'relay.ico'),'rb').read()
_,_,n=struct.unpack('<HHH',ico[:6])
ents=[]
for i in range(n):
    w,h,c,r,pl,bc,size,off=struct.unpack('<BBBBHHII',ico[6+16*i:22+16*i])
    ents.append((w,h,pl,bc,ico[off:off+size]))
def build(machine,reltype,out):
    # resources: type 3 (icons, ids 1..n), type 14 (group icon, id 1)
    blobs=[e[4] for e in ents]
    grp=struct.pack('<HHH',0,1,n)+b''.join(struct.pack('<BBBBHHIH',e[0],e[1],0,0,e[2],e[3],len(e[4]),i+1) for i,e in enumerate(ents))
    blobs.append(grp)
    # layout
    root=16+8*2
    t3=16+8*n
    t14=16+8
    lang_dirs=n+1
    lang_sz=16+8
    base_root=0
    off_t3=root
    off_t14=off_t3+t3
    off_lang=off_t14+t14
    off_data_entries=off_lang+lang_dirs*lang_sz
    off_blobs=off_data_entries+16*(n+1)
    sec=bytearray()
    def dirhdr(nid): return struct.pack('<IIHHHH',0,0,0,0,0,nid)
    sec+=dirhdr(2)
    sec+=struct.pack('<II',3,0x80000000|off_t3)
    sec+=struct.pack('<II',14,0x80000000|off_t14)
    sec+=dirhdr(n)
    for i in range(n): sec+=struct.pack('<II',i+1,0x80000000|(off_lang+i*lang_sz))
    sec+=dirhdr(1)
    sec+=struct.pack('<II',1,0x80000000|(off_lang+n*lang_sz))
    for i in range(n+1):
        sec+=dirhdr(1)
        sec+=struct.pack('<II',0x409,off_data_entries+16*i)
    relocs=[]
    pos=off_blobs; offs=[]
    for b in blobs:
        offs.append(pos); pos+=(len(b)+3)&~3
    for i,b in enumerate(blobs):
        relocs.append(len(sec))
        sec+=struct.pack('<IIII',offs[i],len(b),0,0)
    for b in blobs:
        sec+=b+b'\0'*(((len(b)+3)&~3)-len(b))
    nrel=len(relocs)
    raw_ptr=20+40
    rel_ptr=raw_ptr+len(sec)
    sym_ptr=rel_ptr+10*nrel
    hdr=struct.pack('<HHIIIHH',machine,1,0,sym_ptr,2,0,0)
    sh=struct.pack('<8sIIIIIIHHI',b'.rsrc',0,0,len(sec),raw_ptr,rel_ptr,0,nrel,0,0x40000040)
    rel=b''.join(struct.pack('<IIH',r,0,reltype) for r in relocs)
    sym=struct.pack('<8sIhHBB',b'.rsrc',0,1,0,3,1)+struct.pack('<IHHIHBBBB',len(sec),nrel,0,0,0,0,0,0,0)[:18]
    open(out,'wb').write(hdr+sh+bytes(sec)+rel+sym+struct.pack('<I',4))
for arch,m,rt in (('amd64',0x8664,3),('386',0x14c,7),('arm64',0xAA64,2)):
    build(m,rt,os.path.join(OUT,f'rsrc_windows_{arch}.syso'))
print('ok')
