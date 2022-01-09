//darkMode()

function hideEditStatus() {
    var x = document.getElementsByClassName('hidyform');
    for (let i = 0; i < x.length; i++) {
        if (x[i].style.display === "none") {
            x[i].style.display = "block";
        } else {
            x[i].style.display = "none";
        }
    }
    if (document.getElementById('showbutton').value==="Edit"){
        document.getElementById('showbutton').value="Collapse";
    }else{
        document.getElementById('showbutton').value="Edit";
    }
}

function hideDelete() {
    let x = document.getElementById("hidiv");
    if (x.style.display === "none") {
        x.style.display = "block";
    } else {
        x.style.display = "none";
    }
}

function darkMode() {


    document.body.classList.toggle("dark-mode");
    panelbodies=document.getElementsByClassName('panel-body')
    for (let i=0;i<panelbodies.length; i++) {
        panelbodies.item(i).classList.toggle("dark-mode")
    }
    panelfooters=document.getElementsByClassName('panel-footer')
    for (let i=0;i<panelfooters.length; i++) {
        panelfooters.item(i).classList.toggle("dark-mode")
    }
}

function uploadImage() {
    document.getElementById("postimage").click() ;
}

function addUnfriend(){
    httpRequest = new XMLHttpRequest();
    let url="/addunfriend"
    let friendID = window.location.pathname.split("/").pop();
    httpRequest.onreadystatechange = alertContents;
    httpRequest.open("POST", url,true);
    httpRequest.setRequestHeader('Content-Type', 'application/json');
    httpRequest.send(JSON.stringify({"friendid":friendID}));
}
function alertContents() {
    let addbut=document.getElementById("addfriendbut");
    if (httpRequest.readyState === httpRequest.DONE) {
        if (httpRequest.status === 200) {
            let response = httpRequest.responseText;
            addbut.value = (response==="true")?"Unfriend":"Add Friend";
            addbut.style.backgroundColor=(response==="true")?"firebrick":"#337ab7";
        } else {
            console.log('There was a problem with the request.');
        }
    }
}

function toggleEditBio(){
    let editButton=document.getElementById('edit-bio-btn')
    let formDiv=document.getElementById('update-bio-id')
    if (formDiv.style.display==='none'){
        formDiv.style.display='block'
        editButton.value="Shrink"
    }else {
        formDiv.style.display='none'
        editButton.value="Edit Bio"
    }
}

let page = 1;
function loadMore(url){
    ajaxLoadMore = new XMLHttpRequest();
    page+=1
    url=url+page
    ajaxLoadMore.onreadystatechange=loadContents;
    ajaxLoadMore.responseType="json";
    ajaxLoadMore.open("GET",url,true);
    ajaxLoadMore.send()
}
function loadContents(){
    if (ajaxLoadMore.readyState === ajaxLoadMore.DONE) {
        if (ajaxLoadMore.status === 200) {
            let response = ajaxLoadMore.response; /* responseXML olabilr*/
            //darkMode()
            response.LoadMorePost.forEach(function (element){
                document.getElementById('loadmoredivid').insertAdjacentHTML("beforeend", element);
            })
            //darkMode()
        } else {
            console.log('There was a problem with the request.');
        }
    }
}

function deletePost(url){
    ajaxDelPost = new XMLHttpRequest();
    ajaxDelPost.onreadystatechange=function (){
        if (ajaxDelPost.readyState === ajaxDelPost.DONE) {
            if (ajaxDelPost.status === 200) {
                document.getElementById("dummy-delete").click()
            } else {
                console.log('There was a problem with the request.');
            }
        }
    }
    ajaxDelPost.open("GET",url,true);
    ajaxDelPost.send();
    return true;
}


function searchUser(){

    searchTextBox=document.getElementById('search-box-text');

    if (searchTextBox.value.length>2){
        searchTextAjax = new XMLHttpRequest();
        searchTextAjax.responseType="json";
        searchTextAjax.onreadystatechange = searchUserContent;
        searchTextAjax.open("POST", "/searchuser",true);
        searchTextAjax.setRequestHeader('Content-Type', 'application/json');
        searchTextAjax.send(JSON.stringify({"searchLetters":searchTextBox.value}));
    }
    else{
        document.getElementById('dropdowncontdivid').innerHTML = '';
    }
}
function searchUserContent(){
    if (searchTextAjax.readyState === searchTextAjax.DONE) {
        if (searchTextAjax.status === 200) {
            let response = searchTextAjax.response;

            document.getElementById('dropdowncontdivid').innerHTML = ""
            response.filtered.forEach(function (element){
                document.getElementById('dropdowncontdivid').insertAdjacentHTML("beforeend", element);
            })


        } else {
            console.log('There was a problem with the request.');
        }
    }
}
function searchUserDirectly(){
    document.getElementById('seach-form-id').action = "/user/" + document.getElementById('search-box-text').value;
}
