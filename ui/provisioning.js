let aMm = document.querySelector("a[href='#mnemonic']")
let aHex = document.querySelector("a[href='#hex']")

let cMm = document.getElementById("container-mmkeys")
let cHex = document.getElementById("container-hexkeys")

aMm.addEventListener("click", (el, ev) => {
    aHex.removeAttribute("class");
    aMm.setAttribute("class", "active");
    
    cHex.style = "display: none";
    cMm.style = "";
})

aHex.addEventListener("click", (el, ev) => {
    aMm.removeAttribute("class");
    aHex.setAttribute("class", "active");
    
    cMm.style = "display: none";
    cHex.style = "";
})