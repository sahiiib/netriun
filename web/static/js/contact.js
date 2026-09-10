document.getElementById('contact-form').addEventListener('submit',async event=>{
 event.preventDefault();const form=event.target,button=form.querySelector('button'),result=document.getElementById('contact-result');button.disabled=true;result.textContent='Sending…';
 try{const response=await fetch('/api/contact',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(Object.fromEntries(new FormData(form)))});const data=await response.json();if(!response.ok)throw Error(data.error||'Unable to send your message. Please try again.');result.textContent=data.message;form.reset()}catch(error){result.textContent=error.message}finally{button.disabled=false}
});
