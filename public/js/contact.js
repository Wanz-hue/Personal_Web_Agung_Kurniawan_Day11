function submitdata() {
    let name = document.getElementById("input-name").value;
    let email = document.getElementById("input-email").value;
    let phone = document.getElementById("input-phone").value;
    let subject = document.getElementById("input-subject").value;
    let message = document.getElementById("input-message").value;
    if (name == "") {
      return alert("isikan nama");
    } else if (email == "") {
      return alert("isikan email");
    } else if (phone == "") {
      return alert("isikan nomor telpon");
    } else if (subject == "") {
      return alert("isikan subject");
    } else if (message == "") {
      return alert("isikan message");
    }
    let email_receiver = "agungkurniawan211@gmail.com";
    let a = document.createElement("a");
    
  
    // Kasih Line baru %0D%0A
    a.href = `mailto:${email_receiver}?subject=${subject}&body=Hello my name ${name},2C${subject},2C
      %0D%0A${message}%0D%0A
      you can contact me in my email ${email} or you can call me in my number ${phone}`;
    a.click();
  }
  