package constants

var DefaultHtml = `<title>Loaded</title>`

var V3Html = `<title>V3 Harvester</title>
<body>
<div id="captchaFrame"></div>
<p>Solved: <b id="solvedCounter">0</b></p> 
<button id="manualSubmit" onClick="document.harv.harvest(document.harv.siteKey, document.harv.isEnterprise, document.harv.renderParams, true)">Force Solve</button>
</body>`

var V3Loader = `
(async () => {
	document.harv = {
		counter: 0,
		loaded: false
	};
	
	document.harv.p = new Promise(r => document.harv.pResolve = r);
	document.harv.increment = () => {
		document.harv.counter++;
		document.querySelector('#solvedCounter').textContent = "Solved " + document.harv.counter;
	}	

	window.onLoadCallback = () => document.harv.isEnterprise ? grecaptcha.enterprise.ready(document.harv.pResolve) : grecaptcha.ready(document.harv.pResolve);

	document.harv.harvest = async (siteKey, isEnterprise=false, renderParams={}, isForced=false) => {
		try {
			if(isForced && !document.harv.loaded) {
				return;
			}
	
			if (!document.harv.loaded) {
				document.harv.siteKey = siteKey;
				document.harv.isEnterprise = isEnterprise;
				document.harv.renderParams = renderParams;
	
				const script = document.createElement('script');
				script.src = 'https://www.google.com/recaptcha/' + (isEnterprise ? 'enterprise' : 'api') + '.js?render=' + siteKey + '&onload=onLoadCallback';
				script.setAttribute('async', '');
				script.setAttribute('defer', '');
				document.querySelector('head').appendChild(script);
				await document.harv.p;
				// document.getElementsByClassName('grecaptcha-badge')[0].style.display = 'none';
				document.harv.loaded = true;
			}
	
			const r = isEnterprise ? await grecaptcha.enterprise.execute(siteKey, renderParams) : await grecaptcha.execute(siteKey, renderParams);
			document.harv.increment();
			return r;
		} catch(e) {
			return {"harvest_error": e.message};
		}
	}
})();`