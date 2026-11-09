# Draws the tray/exe icon (relay.ico). Needs: pip install pillow. Run from anywhere.
import struct, io, os
HERE=os.path.dirname(os.path.abspath(__file__)); OUT=os.path.join(HERE,'..')
from PIL import Image, ImageDraw
def base(n):
    S=n*8
    im=Image.new('RGBA',(S,S),(0,0,0,0)); d=ImageDraw.Draw(im)
    d.rounded_rectangle([0,0,S-1,S-1],radius=S*0.22,fill=(24,26,38,255))
    c=S/2; r=S*0.38; w=max(int(S*0.07),2)
    d.arc([c-r,c-r,c+r,c+r],start=-60,end=270,fill=(124,92,255,255),width=w)
    hs=[0.18,0.34,0.5,0.34,0.18]; bw=S*0.07; gap=S*0.045
    tot=len(hs)*bw+(len(hs)-1)*gap; x=c-tot/2
    for i,h in enumerate(hs):
        hh=S*h; d.rounded_rectangle([x,c-hh/2,x+bw,c+hh/2],radius=bw/2,fill=(255,255,255,255) if i==2 else (86,214,196,255)); x+=bw+gap
    d.ellipse([S*0.70,S*0.12,S*0.70+S*0.13,S*0.12+S*0.13],fill=(255,82,82,255))
    return im
def arrow(im):
    S=im.size[0]; d=ImageDraw.Draw(im)
    # bold amber arrow from the bottom-left corner pointing in towards the centre
    import math
    tip=(S*0.50,S*0.50); tail=(S*0.04,S*0.96)
    dx,dy=tip[0]-tail[0],tip[1]-tail[1]; L=math.hypot(dx,dy); ux,uy=dx/L,dy/L; px,py=-uy,ux
    sw=S*0.095; hw=S*0.25; hl=S*0.30
    base=(tip[0]-ux*hl,tip[1]-uy*hl)
    poly=[(tail[0]+px*sw,tail[1]+py*sw),(base[0]+px*sw,base[1]+py*sw),(base[0]+px*hw,base[1]+py*hw),tip,
          (base[0]-px*hw,base[1]-py*hw),(base[0]-px*sw,base[1]-py*sw),(tail[0]-px*sw,tail[1]-py*sw)]
    d.line(poly+[poly[0]],fill=(24,26,38,255),width=int(S*0.07),joint='curve')
    d.polygon(poly,fill=(255,255,255,255),outline=(24,26,38,255),width=int(S*0.035))
    return im
def dib(img):
    w,h=img.size; px=img.convert('RGBA').load()
    rows=b''; mask=b''
    mrow=((w+31)//32)*4
    for y in range(h-1,-1,-1):
        for x in range(w):
            r,g,b,a=px[x,y]; rows+=struct.pack('BBBB',b,g,r,a)
        mask+=b'\0'*mrow
    return struct.pack('<IiiHHIIiiII',40,w,h*2,1,32,0,len(rows)+len(mask),0,0,0,0)+rows+mask
def ico(path,sizes,fn):
    imgs=[fn(n).resize((n,n),Image.LANCZOS) for n in sizes]
    data=[dib(i) for i in imgs]
    out=struct.pack('<HHH',0,1,len(sizes)); off=6+16*len(sizes)
    for n,dd in zip(sizes,data):
        out+=struct.pack('<BBBBHHII',n,n,0,0,1,32,len(dd),off); off+=len(dd)
    open(path,'wb').write(out+b''.join(data))
ico(os.path.join(OUT,'relay.ico'),[16,24,32,48],lambda n: arrow(base(n)))
