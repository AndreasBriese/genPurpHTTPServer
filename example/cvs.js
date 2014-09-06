//      Canvas based photoviewer
//      coded by Andreas Briese
//      (c)2011 eduToolbox@Bri-C GmbH
//      http://eduToolbox.de
// 
//      Vorbereitung:
//      10 jpeg-Bilder mit den Namen 00.jpg .. 09.jpg
//      in einen Unter-Ordner "fotos" legen, der innerhalb des 
//      Ordners liegt, in dem diese html-Datei liegt.
//      use: 
//      put 10 jpeg-Images named 00.jpg .. 09.jpg
//      into a sub-folder called fotos within the folder
//      where this html-file is located.

var photoviewerCanvas = this.document;
var photoviewer = {};
photoviewer.bigBildIdx = 0;
photoviewer.photolist = {};
photoviewer.smallThumbBreite = 60;
photoviewer.smallThumbHoehe = 45;
photoviewer.bigBildBreite = 600;
photoviewer.obenUntenDistMinimum = 35;
photoviewer.bigBildHoehe = 450;
photoviewer.bildZaehler = 0;
photoviewer.autoplayerTimelapse = 2500;

Array.prototype.sum = function() {
	var summe = 0;
	try {
		for (var i = 0; i < this.length; i++) summe += this[i];
	} catch (e) {}
	return summe;
};

Number.prototype.digit = function() {
	return (this >= 0 && this < 10) ? ("0" + this) : this;
};

function Rad(winkel) {
	return (winkel * Math.PI / 180);
}

function drehe(context, dX, dY, winkel) {
	context.save();
	context.setTransform(1, 0, 0, 1, 0, 0);
	context.translate(dX, dY);
	context.rotate(Rad(winkel));
	context.translate(-dX, -dY);
}

function initr() {
	for (var i = 0; i < 10; i++) {
		eval('photoviewer.photolist.photo' + i.digit() + ' = new Image();');
		eval('photoviewer.photolist.photo' + i.digit() + '.addEventListener("load",photoLoaded,false);');
		eval('photoviewer.photolist.photo' + i.digit() + '.src = "fotos/' + i.digit() + '.jpg";');
	}
	this.focus();
	photoviewerCanvas.addEventListener('keydown', keypressd, false);
	photoviewer.autoplayer = setInterval(autoplayNext, photoviewer.autoplayerTimelapse);
	//photoviewerCanvas.getElementById('speed').addEventListener('change',function(){
	//  alert(photoviewerCanvas.getElementById('speed').value);
	//  photoviewer.autoplayerTimelapse = photoviewerCanvas.getElementById('speed').value;
	//  photoviewer.autoplayer = setInterval(autoplayNext, photoviewer.autoplayerTimelapse);
	//  },false);
}

function photoLoaded(evt) {
	photoviewer.bildZaehler++;
	if (photoviewer.bildZaehler == 10)
		renderSmall();
	renderBig(photoviewer.photolist.photo00);
}

function checkCol(imageData) {
	var colsum = 0;
	var minusFaktor = 2;
	var retVal = [];
	for (var i = 1; i < imageData.data.length; i += 4) {
		if ((imageData.data[i] + imageData.data[i + 1] + imageData.data[i + 2]) > colsum) {
			colsum = imageData.data[i] + imageData.data[i + 1] + imageData.data[i + 2];
			retVal = [];
			retVal.push(Math.floor(imageData.data[i] / minusFaktor));
			retVal.push(Math.floor(imageData.data[i + 1] / minusFaktor));
			retVal.push(Math.floor(imageData.data[i + 2] / minusFaktor));
		}
	}
	return retVal;
}

function renderSmall() {
	var smallBackground = "rgba(255,255,255,0.7)";
	var small = photoviewerCanvas.getElementById('small');
	smallCtx = small.getContext('2d');
	smallCtx.fillStyle = smallBackground;
	smallCtx.fillRect(0, 0, 10 * photoviewer.smallThumbBreite, photoviewer.smallThumbHoehe);
	smallCtx.strokeStyle = smallBackground;
	smallCtx.lineWidth = 1;
	for (var i = 0; i < photoviewer.bildZaehler; i++) {
		eval('var bild = photoviewer.photolist.photo' + i.digit());
		var diesesBildHoehe = Math.floor(bild.height * photoviewer.smallThumbBreite / bild.width);
		if (diesesBildHoehe > (photoviewer.smallThumbHoehe - 6)) diesesBildHoehe = photoviewer.smallThumbHoehe - 6;
		var OberLinienHoehe = Math.floor(0.5 * (photoviewer.smallThumbHoehe - diesesBildHoehe));
		if (OberLinienHoehe < 2) OberLinienHoehe = 2;
		smallCtx.drawImage(bild, 0, 0, bild.width, bild.height, i * photoviewer.smallThumbBreite, OberLinienHoehe, photoviewer.smallThumbBreite, diesesBildHoehe);
		smallCtx.strokeRect((0.5 + i * photoviewer.smallThumbBreite), OberLinienHoehe, photoviewer.smallThumbBreite - 1, diesesBildHoehe);
	}
	small.style.top = (photoviewer.bigBildHoehe + 1) + "px";
	small.style.left = "0px";
	small.style.display = 'block';
	photoviewerCanvas.addEventListener('mousemove', mouseOverSmall, false);
}

function mouseOverSmall(evt) {
	mouseX = evt.clientX;
	mouseY = evt.clientY;
	if (mouseY > photoviewer.bigBildHoehe && mouseY <= photoviewer.bigBildHoehe + photoviewer.smallThumbHoehe) {
		photoviewerCanvas.getElementById('small').style.opacity = '0.5';
		if (photoviewer.autoplayer) clearInterval(photoviewer.autoplayer);
		if (Math.floor(mouseX / photoviewer.smallThumbBreite) != photoviewer.bigBildIdx) {
			photoviewer.bigBildIdx = Math.floor(mouseX / photoviewer.smallThumbBreite);
			eval('var bild = photoviewer.photolist.photo' + photoviewer.bigBildIdx.digit());
			if (bild) {
				renderBig(bild);
			}
		}
	} else {
		var small = photoviewerCanvas.getElementById('small');
		if (small.style.opacity != '1') small.style.opacity = '1';
	}
}

function renderBig(bild) {
	try{
		var obenUntenDistMinimum = photoviewer.obenUntenDistMinimum;
		var bigTextOben1 = "Screenshots";
		var bigTextOben2 = "";
		var bigTextUnten = "";
		var bigBackground = ["rgba(80,95,95,", "rgba(30,50,30,", "rgba(70,20,20,", "rgba(200,200,200,", "rgba(130,130,130,", "rgba(90,30,30,", "rgba(30,50,30,", "rgba(200,200,200,", "rgba(30,30,30,", "rgba(90,30,30,"];
		var big = photoviewerCanvas.getElementById('big');
		bigCtx = big.getContext('2d');
		var diesesBildHoehe = Math.floor(bild.height * photoviewer.bigBildBreite / bild.width);
		if (diesesBildHoehe > photoviewer.bigBildHoehe){ 
			diesesBildHoehe = photoviewer.bigBildHoehe;
		}
		var OberLinienHoehe = Math.floor(0.5 * (photoviewer.bigBildHoehe - diesesBildHoehe));
		bigCtx.fillStyle = "rgba(200,200,200,1)";
		bigCtx.fillRect(0, 0, photoviewer.bigBildBreite, photoviewer.bigBildHoehe);
		bigCtx.drawImage(bild, 0, 0, bild.width, bild.height, 0, OberLinienHoehe, photoviewer.bigBildBreite, diesesBildHoehe);
		var bigImageData = bigCtx.createImageData(10, diesesBildHoehe);
		bigImageData = bigCtx.getImageData(0, OberLinienHoehe, 10, diesesBildHoehe);
		var sideColsL = checkCol(bigImageData);
		bigImageData = bigCtx.getImageData(photoviewer.bigBildBreite - 10, OberLinienHoehe, 10, diesesBildHoehe);
		var sideColsR = checkCol(bigImageData);
		var sideCols = (sideColsL.sum() > sideColsR.sum()) ? sideColsL : sideColsR;
		bigCtx.strokeStyle = "rgba(" + sideCols[1] + "," + sideCols[2] + "," + sideCols[0] + ",0.4)";
		bigCtx.lineWidth = OberLinienHoehe;
		bigCtx.strokeRect(0.5, OberLinienHoehe, photoviewer.bigBildBreite - 1, diesesBildHoehe - 1);
		bigCtx.fillStyle = "rgba(10,10,10,1)";
		bigCtx.fillRect(0, 0, photoviewer.bigBildBreite, OberLinienHoehe);
		//bigCtx.fillRect(0, (diesesBildHoehe + OberLinienHoehe), photoviewer.bigBildBreite, OberLinienHoehe);
		bigCtx.fillStyle = "rgba(" + sideCols[1] + "," + sideCols[2] + "," + sideCols[0] + ",0.7)";
		bigCtx.fillRect(0, 0, photoviewer.bigBildBreite, OberLinienHoehe);
		//bigCtx.fillRect(0, (diesesBildHoehe + OberLinienHoehe), photoviewer.bigBildBreite, OberLinienHoehe);
		bigCtx.fillStyle = "rgba(255,255,255,1)";
		//bigCtx.font = "normal bold 24px fantasy";
		bigCtx.font = "normal bold 20px sans-serif";
		bigCtx.textBaseline = "bottom";
		bigCtx.textAlign = "start";
		bigCtx.fillText(bigTextOben1, 7, 25); //
		bigCtx.font = "normal bold 20px sans-serif";
		bigCtx.textBaseline = "bottom";
		bigCtx.textAlign = "end";
		bigCtx.fillText(bigTextOben2, (photoviewer.bigBildBreite - 7), 25); //obenUntenDistMinimum); //OberLinienHoehe );
		//UNTEN-Txt
		//bigCtx.fillStyle = "rgba(255,255,255,0.6)";
		//bigCtx.font = "normal bold 14px sans-serif";
		//bigCtx.fillText(bigTextUnten, (photoviewer.bigBildBreite-6), photoviewer.bigBildHoehe-10);

		bigCtx.setTransform(1, 0, 0, 1, 0, 0);
		var copyrightTxt = '(C) 2011 eduToolbox@Bri-C';
		var copyrightX = photoviewer.bigBildBreite - 3;
		var copyrightY = photoviewer.bigBildHoehe - OberLinienHoehe - bigCtx.measureText(copyrightTxt).width;
		drehe(bigCtx, copyrightX, copyrightY, 270);
		bigCtx.fillStyle = "#aaa";
		bigCtx.font = "normal bold 10px sans-serif";
		bigCtx.fillText(copyrightTxt, copyrightX, copyrightY);
		bigCtx.setTransform(1, 0, 0, 1, 0, 0);

		big.style.top = '0px';
		big.style.left = "0px";
		big.style.display = 'block';
	} catch(e){}
}

function keypressd(evt) {
	switch (evt.keyCode) {
		case 37:
			if (photoviewer.autoplayer) clearInterval(photoviewer.autoplayer);
			photoviewer.bigBildIdx = (photoviewer.bigBildIdx > 0) ? (photoviewer.bigBildIdx - 1) : photoviewer.bildZaehler - 1;
			renderBig(eval('photoviewer.photolist.photo' + photoviewer.bigBildIdx.digit()));
			break;
		case 39:
			if (photoviewer.autoplayer) clearInterval(photoviewer.autoplayer);
			photoviewer.bigBildIdx = (photoviewer.bigBildIdx < photoviewer.bildZaehler - 1) ? (photoviewer.bigBildIdx + 1) : 0;
			renderBig(eval('photoviewer.photolist.photo' + photoviewer.bigBildIdx.digit()));
			break;
	}
	photoviewerCanvas.getElementById('small').style.opacity = '0.5';
}

function canvasCheckr() {
	if (!photoviewerCanvas.createElement('canvas').getContext) {
		var cp = photoviewerCanvas.createElement('div');
		cp.id = "copyrights";
		cp.firstChild.innerHTML = '(C) 2014 eduToolbox@Bri-C';
		photoviewerCanvas.body.appendChild(cp);
		if (window.navigator.appName.find('MSIE') != -1) makeIEWindow();
	} else initr();
}

function autoplayNext() {
	photoviewer.bigBildIdx = (photoviewer.bigBildIdx < photoviewer.bildZaehler - 1) ? (photoviewer.bigBildIdx + 1) : 0;
	renderBig(eval('photoviewer.photolist.photo' + photoviewer.bigBildIdx.digit()));
	photoviewerCanvas.getElementById('small').style.opacity = '0.5';
}

function makeIEWindow() {
	while (documnet.body.childNodes.length > 1)
		document.body.removeNode(document.body.firstChild);
	var bildFeld = photoviewerCanvas.createElement('div');
	bildFeld.id = "bildFeld";
	bildFeld.style.width = "800px";
	bildFeld.style.height = "600px";

	photoviewerCanvas.body.appendChild(cp);

}

photoviewerCanvas.onload = canvasCheckr();
