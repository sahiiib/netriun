#!/usr/bin/env python3
"""Run against the disposable local editor database/server only."""
import copy
import http.cookiejar
import json
import os
from urllib.request import Request, build_opener, HTTPCookieProcessor
from urllib.error import HTTPError
from urllib.parse import urlparse
origin = os.environ.get('TEST_ORIGIN', 'http://localhost:18088')
if urlparse(origin).hostname not in ('localhost','127.0.0.1'):
    raise SystemExit('Refusing to write test content outside localhost')
jar = http.cookiejar.CookieJar()
client = build_opener(HTTPCookieProcessor(jar))
csrf = ''
def req(path, method='GET', data=None, status=200):
    request = Request(origin+path, method=method, data=None if data is None else json.dumps(data).encode(), headers={'Content-Type':'application/json', 'Origin':origin, 'X-CSRF-Token':csrf})
    try:
        result = client.open(request)
    except HTTPError as error:
        result = error
    body = result.read().decode()
    assert result.code == status, (path, result.code, body)
    return json.loads(body) if 'application/json' in result.headers.get('Content-Type','') else body
api='/neditport2065/api'
req(api+'/content',status=401)
session=req(api+'/login','POST',{'username':'admin','password':os.environ.get('TEST_ADMIN_PASSWORD','local-preview-password-only')})
csrf=session['csrf']
state=req(api+'/content');original=copy.deepcopy(state['content']);c=state['content'];revision=state['revision']
try:
    c['news']=[{'id':'test-news','name':{'en':'Test news <script>alert(1)</script>'},'description':{'en':'News body from the editor.'},'status':'draft','image':'','url':'','date':'2026-09-10'}]
    c['partners']=[{'id':'test-partner','name':{'en':'Test Partner'},'description':{'en':'A local-only partner.'},'status':'draft','image':'','url':'https://example.com','date':''}]
    c['about']['lead']['en']='Local end-to-end About text'
    revision=req(api+'/content','PUT',{'content':c,'revision':revision})['revision']
    assert 'Test news' not in req('/news')
    assert 'Test Partner' not in req('/')
    assert 'Local end-to-end About text' in req('/about')
    c['news'][0]['status']='published';c['partners'][0]['status']='published'
    revision=req(api+'/content','PUT',{'content':c,'revision':revision})['revision']
    news=req('/news');assert 'News body from the editor.' in news and '&lt;script&gt;' in news and '<script>alert(1)</script>' not in news
    assert 'Test Partner' in req('/')
    for path in ['/','/products']:
        body=req(path);assert body.count('data-i18n="products.comingSoon"')==3 and 'href="https://nexus.netriun.com/"' in body
    req('/api/contact','POST',{'name':'Local Test','email':'test@example.com','company':'Example','topic':'support','message':'A local test message only.','website':''})
    assert any(m['email']=='test@example.com' for m in req(api+'/messages'))
    print('PASS: login, draft visibility, publication, escaping, About, products and contact inbox')
finally:
    req(api+'/content','PUT',{'content':original,'revision':revision})
    req(api+'/logout','POST',{})
req(api+'/content',status=401)
print('PASS: sign-out revokes session; local content restored')
