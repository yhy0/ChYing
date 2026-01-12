package javascript

import (
    "fmt"
    "github.com/tdewolff/parse/v2"
    "github.com/tdewolff/parse/v2/js"
    "testing"
)

func TestJqueryWalker(t *testing.T) {
    source := `
jQuery.getScript("https://dev.jquery.com/view/trunk/plugins/color/jquery.color.js", function(){
  $("#go").click(function(){
    $(".block").animate( { backgroundColor: 'pink' }, 1000)
      .animate( { backgroundColor: 'blue' }, 1000);
  });
});

$.post("test.php", { "func": "getNameAndTime" },
  function(data){
  	alert(data.name);
  	console.log(data.time);
  }, "json");

jQuery.post({url: "/example"});
$("#feeds").load("feeds.html");
$("button").get("test2222.html");
a.get("nouse");
Cms.channelViewCount = function(base, channelId, viewId) {
	viewId = viewId || "views";
	$.getJSON(base + "/channel_view.jspx", {
		channelId : channelId
	});
}
Cms = {};
/**
 * 娴忚娆℃暟
 */
Cms.viewCount = function(base, contentId, viewId, commentId, downloadId, upId,
		downId) {
	viewId = viewId || "views";
	commentId = commentId || "comments";
	downloadId = downloadId || "downloads";
	upId = upId || "ups";
	downId = downId || "downs";
	$.getJSON(base + "/content_view.jspx", {
		contentId : contentId
	}, function(data) {
		if (data.length > 0) {
			$("#" + viewId).text(data[0]);
			$("#" + commentId).text(data[1]);
			$("#" + downloadId).text(data[2]);
			$("#" + upId).text(data[3]);
			$("#" + downId).text(data[4]);
		}
	});
}
Cms.channelViewCount = function(base, channelId, viewId) {
	viewId = viewId || "views";
	$.getJSON(base + "/channel_view.jspx", {
		channelId : channelId
	});
}
/**
 * 绔欑偣娴侀噺缁熻
 */
Cms.siteFlow = function(base,page, referer,flowSwitch,
		pvId,visitorId,dayPvId, dayVisitorId,
		weekPvId,weekVisitorId,monthPvId,monthVisitorId) {
	pvId = pvId || "pv";
	visitorId = visitorId || "visitor";
	dayPvId=dayPvId || "dayPv";
	dayVisitorId=dayVisitorId || "dayVisitor";
	weekPvId=weekPvId || "weekPv";
	weekVisitorId=weekVisitorId || "weekVisitor";
	monthPvId=monthPvId || "monthPv";
	monthVisitorId=monthVisitorId || "monthVisitor";
	flowSwitch=flowSwitch||"true";
	if(flowSwitch=="true"){
		$.getJSON(base + "/flow_statistic.jspx", {
			page : page,
			referer : referer
		}, function(data) {
			if (data.length > 0) {
				$("#" + pvId).text(data[0]);
				$("#" + visitorId).text(data[1]);
				$("#" + dayPvId).text(data[2]);
				$("#" + dayVisitorId).text(data[3]);
				$("#" + weekPvId).text(data[4]);
				$("#" + weekVisitorId).text(data[5]);
				$("#" + monthPvId).text(data[6]);
				$("#" + monthVisitorId).text(data[7]);
			}
		});
	}
}
/**
 * 鎴愬姛杩斿洖true锛屽け璐ヨ繑鍥瀎alse銆�
 */
Cms.up = function(base, contentId, origValue, upId) {
	upId = upId || "ups";
	var updown = $.cookie("_cms_updown_" + contentId);
	if (updown) {
		return false;
	}
	$.cookie("_cms_updown_" + contentId, "1");
	$.get(base + "/content_up.jspx", {
		"contentId" : contentId
	}, function(data) {
		$("#" + upId).text(origValue + 1);
	});
	return true;
}
/**
 * 鎴愬姛杩斿洖true锛屽け璐ヨ繑鍥瀎alse銆�
 */
Cms.down = function(base, contentId, origValue, downId) {
	downId = downId || "downs";
	var updown = $.cookie("_cms_updown_" + contentId);
	if (updown) {
		return false;
	}
	$.cookie("_cms_updown_" + contentId, "1");
	$.get(base + "/content_down.jspx", {
		contentId : contentId
	}, function(data) {
		$("#" + downId).text(origValue + 1);
	});
	return true;
}
/**
 * 鑾峰彇璇勫垎閫夐」鎶曠エ鏁�
 */
Cms.scoreCount = function(base, contentId,itemPrefix) {
	itemPrefix=itemPrefix||"score-item-";
	$.getJSON(base + "/content_score_items.jspx", {
		contentId : contentId
	}, function(data) {
			$("span[id^='"+itemPrefix+"']").each(function(){
				var itemId=$(this).prop("id").split(itemPrefix)[1];
				$(this).text(data.result[itemId]);
			});
	});
}
/**
 * 鎴愬姛杩斿洖true锛屽け璐ヨ繑鍥瀎alse銆�
 */
Cms.score = function(base, contentId,itemId,itemPrefix) {
	itemPrefix=itemPrefix||"score-item-";
	var score = $.cookie("_cms_score_" + contentId);
	if (score) {
		return false;
	}
	$.cookie("_cms_score_" + contentId, "1");
	$.get(base + "/content_score.jspx", {
		"contentId" : contentId,
		"itemId":itemId
	}, function(data) {
		if(data.succ){
			$("#"+itemPrefix + itemId).text(data.count);
		}
	});
	return true;
}
/**
 * 鑾峰彇闄勪欢鍦板潃
 */
Cms.attachment = function(base, contentId, n, prefix) {
	$.get(base + "/attachment_url.jspx", {
		"cid" : contentId,
		"n" : n
	}, function(data) {
		var url;
		for (var i = 0;i < n; i++) {
			url = base + "/attachment.jspx?cid=" + contentId + "&i=" + i
					+ data[i];
			$("#" + prefix + i).attr("href", url);
		}
	}, "json");
}
/**
 * 鎻愪氦璇勮
 */
Cms.comment = function(callback, form) {
	form = form || "commentForm";
	$("#" + form).validate( {
		submitHandler : function(form) {
			$(form).ajaxSubmit( {
				"success" : callback,
				"dataType" : "json"
			});
		}
	});
}
/**
 * 鑾峰彇璇勮鍒楄〃
 * 
 * @param siteId
 * @param contentId
 * @param greatTo
 * @param recommend
 * @param orderBy
 * @param count
 */
Cms.commentList = function(base, c, options) {
	c = c || "commentListDiv";
	$("#" + c).load(base + "/comment_list.jspx", options);
}
Cms.commentListMore = function(base, c, options) {
	c = c || "commentListDiv";
	$("#" + c).load(base + "/comment_list.jspx", options);
	$('#commentDialog').dialog('open');
}
/**
 * 璇勮椤�
 */
Cms.commentUp = function(base, commentId, origValue, upId) {
	upId = upId || "commentups";
	var updown = $.cookie("_cms_comment_updown_" + commentId);
	if (updown) {
		return false;
	}
	$.cookie("_cms_comment_updown_" + commentId, "1");
	$.get(base + "/comment_up.jspx", {
		"commentId" : commentId
	}, function(data) {
		$("#" + upId).text(origValue + 1);
	});
	return true;
}
/**
 * 璇勮韪�
 */
Cms.commentDown = function(base, commentId, origValue, downId) {
	downId = downId || "commentdowns";
	var updown = $.cookie("_cms_comment_updown_" + commentId);
	if (updown) {
		return false;
	}
	$.cookie("_cms_comment_updown_" + commentId, "1");
	$.get(base + "/comment_down.jspx", {
		commentId : commentId
	}, function(data) {
		$("#" + downId).text(origValue + 1);
	});
	return true;
}
/**
 * 璇勮杈撳叆妗�
 */
Cms.commentInputCsi = function(base,commentInputCsiDiv, contentId,commemtId) {
	commentInputCsiDiv = commentInputCsiDiv || "commentInputCsiDiv";
	$("#"+commentInputCsiDiv).load(base+"/comment_input_csi.jspx?contentId="+contentId+"&commemtId="+commemtId);
}
Cms.commentInputLoad= function(base,commentInputCsiPrefix,commentInputCsiDiv,contentId,commemtId) {
	$("div[id^='"+commentInputCsiPrefix+"']").html("");
	Cms.commentInputCsi(base,commentInputCsiDiv,contentId,commemtId);
}
/**
 * 鏄惁鏄井淇℃墦寮€
 */
Cms.isOpenInWeiXin = function() {
	var ua = navigator.userAgent.toLowerCase();
    if(ua.match(/MicroMessenger/i)=="micromessenger") {
        return true;
     } else {
        return false;
    }
}
/**
 * 瀹㈡埛绔寘鍚櫥褰�
 */
Cms.loginCsi = function(base, c, options) {
	c = c || "loginCsiDiv";
	$("#" + c).load(base + "/login_csi.jspx", options);
}
/**
 * 鍚戜笂婊氬姩js绫�
 */
Cms.UpRoller = function(rid, speed, isSleep, sleepTime, rollRows, rollSpan,
		unitHight) {
	this.speed = speed;
	this.rid = rid;
	this.isSleep = isSleep;
	this.sleepTime = sleepTime;
	this.rollRows = rollRows;
	this.rollSpan = rollSpan;
	this.unitHight = unitHight;
	this.proll = $('#roll-' + rid);
	this.prollOrig = $('#roll-orig-' + rid);
	this.prollCopy = $('#roll-copy-' + rid);
	// this.prollLine = $('#p-roll-line-'+rid);
	this.sleepCount = 0;
	this.prollCopy[0].innerHTML = this.prollOrig[0].innerHTML;
	var o = this;
	this.pevent = setInterval(function() {
		o.roll.call(o)
	}, this.speed);
}
Cms.UpRoller.prototype.roll = function() {
	if (this.proll[0].scrollTop > this.prollCopy[0].offsetHeight) {
		this.proll[0].scrollTop = this.rollSpan + 1;
	} else {
		if (this.proll[0].scrollTop % (this.unitHight * this.rollRows) == 0
				&& this.sleepCount <= this.sleepTime && this.isSleep) {
			this.sleepCount++;
			if (this.sleepCount >= this.sleepTime) {
				this.sleepCount = 0;
				this.proll[0].scrollTop += this.rollSpan;
			}
		} else {
			var modCount = (this.proll[0].scrollTop + this.rollSpan)
					% (this.unitHight * this.rollRows);
			if (modCount < this.rollSpan) {
				this.proll[0].scrollTop += this.rollSpan - modCount;
			} else {
				this.proll[0].scrollTop += this.rollSpan;
			}
		}
	}
}
Cms.LeftRoller = function(rid, speed, rollSpan) {
	this.rid = rid;
	this.speed = speed;
	this.rollSpan = rollSpan;
	this.proll = $('#roll-' + rid);
	this.prollOrig = $('#roll-orig-' + rid);
	this.prollCopy = $('#roll-copy-' + rid);
	this.prollCopy[0].innerHTML = this.prollOrig[0].innerHTML;
	var o = this;
	this.pevent = setInterval(function() {
		o.roll.call(o)
	}, this.speed);
}
Cms.LeftRoller.prototype.roll = function() {
	if (this.proll[0].scrollLeft > this.prollCopy[0].offsetWidth) {
		this.proll[0].scrollLeft = this.rollSpan + 1;
	} else {
		this.proll[0].scrollLeft += this.rollSpan;
	}
}
/**
 * 鏀惰棌淇℃伅
 */
Cms.collect = function(base, cId, operate,showSpanId,hideSpanId) {
	$.post(base + "/member/collect.jspx", {
		"cId" : cId,
		"operate" : operate
	}, function(data) {
		if(data.result){
			if(operate==1){
				alert("鏀惰棌鎴愬姛锛�");
				$("#"+showSpanId).show();
				$("#"+hideSpanId).hide();
			}else{
				alert("鍙栨秷鏀惰棌鎴愬姛锛�");
				$("#"+showSpanId).hide();
				$("#"+hideSpanId).show();
			}
		}else{
			alert("璇峰厛鐧诲綍");
		}
	}, "json");
}
/**
 * 鍒楄〃鍙栨秷鏀惰棌淇℃伅
 */
Cms.cmsCollect = function(base, cId, operate) {
	$.post(base + "/member/collect.jspx", {
		"cId" : cId,
		"operate" : operate
	}, function(data) {
		if(data.result){
			if(operate==1){
				alert("鏀惰棌鎴愬姛锛�");
			}else{
				alert("鍙栨秷鏀惰棌鎴愬姛锛�");
				$("#tr_"+cId).remove();
			}
		}else{
			alert("璇峰厛鐧诲綍");
		}
	}, "json");
}
/**
 * 妫€娴嬫槸鍚﹀凡缁忔敹钘忎俊鎭�
 */
Cms.collectexist = function(base, cId,showSpanId,hideSpanId) {
	$.post(base + "/member/collect_exist.jspx", {
		"cId" : cId
	}, function(data) {
		if(data.result){
			$("#"+showSpanId).show();
			$("#"+hideSpanId).hide();
		}else{
			$("#"+showSpanId).hide();
			$("#"+hideSpanId).show();
		}
	}, "json");
}

/**
 * 鐢宠鑱屼綅淇℃伅
 */
Cms.jobApply = function(base, cId) {
	$.post(base + "/member/jobapply.jspx", {
		"cId" : cId
	}, function(data) {
		if(data.result==-1){
			alert("璇峰厛鐧诲綍");
			location.href=base+"/login.jspx";
		}else if(data.result==-2){
			alert("鑱屼綅id涓嶈兘涓虹┖");
		}else if(data.result==-3){
			alert("鏈壘鍒拌鑱屼綅");
		}else if(data.result==-4){
			alert("鎮ㄨ繕娌℃湁鍒涘缓绠€鍘嗭紝璇峰厛瀹屽杽绠€鍘�");
		}else if(data.result==0){
			alert("鎮ㄤ粖澶╁凡缁忕敵璇蜂簡璇ヨ亴浣�!");
		}else if(data.result==1){
			alert("鎴愬姛鐢宠浜嗚鑱屼綅!");
		}
	}, "json");
}
Cms.loginSSO=function(base){
	var username=$.cookie('username');
	var sessionId=$.cookie('JSESSIONID');
	var ssoLogout=$.cookie('sso_logout');
	if(username!=null){
		if(sessionId!=null||(ssoLogout!=null&&ssoLogout=="true")){
			$.post(base+"/sso/login.jspx", {
				username:username,
				sessionId:sessionId,
				ssoLogout:ssoLogout
			}, function(data) {
					if(data.result=="login"||data.result=="logout"){
						location.reload();
					}
			}, "json");
		}
	}
}
Cms.loginAdmin=function(base){
	var sessionKey=localStorage.getItem("sessionKey");
	if(sessionKey==null||sessionKey==""){
		$.post(base+"/adminLogin.jspx", {
		}, function(data) {
			if(data.sessionKey!=""){
				localStorage.setItem("sessionKey", data.sessionKey); 
				localStorage.setItem("userName", data.userName); 
			}
		}, "json");
	}
}
Cms.logoutAdmin=function(base){
	var sessionKey=localStorage.getItem("sessionKey");
	var userName=localStorage.getItem("userName");
	if(sessionKey!=null&&sessionKey!=""&&userName!=""){
		$.post(base+"/adminLogout.jspx", {
			userName:userName,
			sessionKey:sessionKey
		}, function(data) {
		}, "json");
		localStorage.removeItem("sessionKey");
		localStorage.removeItem("userName");
	}
}
Cms.checkPerm = function(base, contentId) {
	$.getJSON(base + "/page_checkperm.jspx", {
		contentId : contentId
	}, function(data) {
		if (data==3) {
			alert("璇峰厛鐧诲綍");
			location.href=base+"/user_no_login.jspx";
		}else if(data==4){
			location.href=base+"/group_forbidden.jspx";
		}else if(data==5){
			location.href=base+"/content/buy.jspx?contentId="+contentId;
		}
	});
}
Cms.collectCsi = function(base,collectCsiDiv, tpl, contentId) {
	collectCsiDiv = collectCsiDiv || "collectCsiDiv";
	$("#"+collectCsiDiv).load(base+"/csi_custom.jspx?tpl="+tpl+"&cId="+contentId);
}
Cms.getCookie=function getCookie(c_name){
	if (document.cookie.length>0)
	  {
	  	c_start=document.cookie.lastIndexOf(c_name + "=");
		  if (c_start!=-1)
		    { 
			    c_start=c_start + c_name.length+1;
			    c_end=document.cookie.indexOf(";",c_start);
			    if (c_end==-1){
			    	c_end=document.cookie.length;
			    } 
			    return unescape(document.cookie.substring(c_start,c_end));
		    } 
		  }
	return "";
}
Cms.MobileUA=function(){
	var ua = navigator.userAgent.toLowerCase();  
    var mua = {  
        IOS: /ipod|iphone|ipad/.test(ua), //iOS  
        IPHONE: /iphone/.test(ua), //iPhone  
        IPAD: /ipad/.test(ua), //iPad  
        ANDROID: /android/.test(ua), //Android Device  
        WINDOWS: /windows/.test(ua), //Windows Device  
        TOUCH_DEVICE: ('ontouchstart' in window) || /touch/.test(ua), //Touch Device  
        MOBILE: /mobile/.test(ua), //Mobile Device (iPad)  
        ANDROID_TABLET: false, //Android Tablet  
        WINDOWS_TABLET: false, //Windows Tablet  
        TABLET: false, //Tablet (iPad, Android, Windows)  
        SMART_PHONE: false //Smart Phone (iPhone, Android)  
    };  
    mua.ANDROID_TABLET = mua.ANDROID && !mua.MOBILE;  
    mua.WINDOWS_TABLET = mua.WINDOWS && /tablet/.test(ua);  
    mua.TABLET = mua.IPAD || mua.ANDROID_TABLET || mua.WINDOWS_TABLET;  
    mua.SMART_PHONE = mua.MOBILE && !mua.TABLET;  
    return mua;  
}

`
    ast, err := js.Parse(parse.NewInputString(source), js.Options{})
    if err != nil {
        t.Fatal(err)
    }
    js.Walk(&walker{}, ast)
}

func TestAjaxWalker(t *testing.T) {
    source := `httpRequest.open('GET', 'http://www.example.org/some.file', true);
	httpRequest.send();
	httpRequest.open('GET', 'http://www.example.org/some.file', true,"testuser","testPass");`
    ast, err := js.Parse(parse.NewInputString(source), js.Options{})
    if err != nil {
        t.Fatal(err)
    }
    js.Walk(&walker{}, ast)
}

func TestAngularWalker(t *testing.T) {
    // build from https://github.com/didinj/angular6-httpclient-example/blob/6d24603db2d1715da28553860f32dcaef0acf5a3/src/app/rest.service.ts
    source := `var RestService = /** @class */ (function () {
    function RestService(http) {
        this.http = http;
    }
    RestService.prototype.extractData = function (res) {
        var body = res;
        return body || {};
    };
    RestService.prototype.getProducts = function () {
        return this.http.get(endpoint + 'products').pipe(Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_3__["map"])(this.extractData));
    };
    RestService.prototype.getProduct = function (id) {
        return this.http.get(endpoint + 'products/' + id).pipe(Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_3__["map"])(this.extractData));
    };
    RestService.prototype.addProduct = function (product) {
        console.log(product);
        return this.http.post(endpoint + 'products', JSON.stringify(product), httpOptions).pipe(Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_3__["tap"])(function (product) { return console.log("added product w/ id=" + product.id); }), Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_3__["catchError"])(this.handleError('addProduct')));
    };
    RestService.prototype.updateProduct = function (id, product) {
        return this.http.put(endpoint + 'products/' + id, JSON.stringify(product), httpOptions).pipe(Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_3__["tap"])(function (_) { return console.log("updated product id=" + id); }), Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_3__["catchError"])(this.handleError('updateProduct')));
    };
    RestService.prototype.deleteProduct = function (id) {
        return this.http.delete(endpoint + 'products/' + id, httpOptions).pipe(Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_3__["tap"])(function (_) { return console.log("deleted product id=" + id); }), Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_3__["catchError"])(this.handleError('deleteProduct')));
    };
    RestService.prototype.handleError = function (operation, result) {
        if (operation === void 0) { operation = 'operation'; }
        return function (error) {
            // TODO: send the error to remote logging infrastructure
            console.error(error); // log to console instead
            // TODO: better job of transforming error for user consumption
            console.log(operation + " failed: " + error.message);
            // Let the app keep running by returning an empty result.
            return Object(rxjs__WEBPACK_IMPORTED_MODULE_2__["of"])(result);
        };
    };
    RestService = __decorate([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_0__["Injectable"])({
            providedIn: 'root'
        }),
        __metadata("design:paramtypes", [_angular_common_http__WEBPACK_IMPORTED_MODULE_1__["HttpClient"]])
    ], RestService);
    return RestService;
}());

httpClient.jsonp(this.heroesURL, callback);
httpClient.patch(url, {name: heroName}, httpOptions).pipe(catchError(this.handleError('patchHero')));
httpClient.request('GET', this.heroesUrl + '?' + 'name=term', {responseType:'json'});
httpClient.request('GET', this.heroesUrl, {responseType:'json'});
httpClient.request('GET', '/test', {responseType:'json'});
httpClient.jsonp("http://baidu.com", callback);
httpClient.jsonp(this.heroesUrl + '?' + 'name=term', callback);
httpClient.jsonp(this.heroesUrl, callback);
`
    ast, err := js.Parse(parse.NewInputString(source), js.Options{})
    if err != nil {
        t.Fatal(err)
    }
    js.Walk(&walker{}, ast)
}

func TestAxiosWalker(t *testing.T) {
    source := `
axios.get("http://citi.hoertlehner.com:8821/api/apartements")
    .then(function(data) {
        console.log("PERFECt.. all ok!");
   console.log("number of apartements: " + data.length);
        console.log(data);
    })
    .catch(function(err) {
        console.log("ERR");
        console.log(err);
    })`
    ast, err := js.Parse(parse.NewInputString(source), js.Options{})
    if err != nil {
        t.Fatal(err)
    }
    js.Walk(&walker{}, ast)
}

func TestFetchWalker(t *testing.T) {
    source := `fetch(url, {
  method: "POST",
  body: JSON.stringify(data),
  headers: {
    "Content-Type": "application/json"
  },
  credentials: "same-origin"
}).then(function(response) {
  response.status     //=> number 100–599
  response.statusText //=> String
  response.headers    //=> Headers
  response.url        //=> String

  return response.text()
}, function(error) {
  error.message //=> String
})

async function withFetch() {
  const res = await fetch('https://jsonplaceholder.typicode.com/posts')
  const json = await res.json()

  return json
}`
    ast, err := js.Parse(parse.NewInputString(source), js.Options{})
    if err != nil {
        t.Fatal(err)
    }
    js.Walk(&walker{}, ast)
}

func TestRouteWalker(t *testing.T) {
    source := `Routes = [
	 {
	   path: 'product', component: ProductComponent, children: [{
	     path: 'detail', component: ProductDetailComponent
	   }, {
	     path: '', redirectTo: 'detail', pathMatch: 'full'
	   }]
	 }
	];

	Routes = [
	 { path: 'home', component: HomeComponent },
	 { path: '', redirectTo: 'home', pathMatch: 'full' },
	 { path: 'news', component: NewsComponent },
	 { path: 'product', component: ProductComponent },
	 { path: '**', component: PagenotfoundComponent },
	];
`
    ast, err := js.Parse(parse.NewInputString(source), js.Options{})
    if err != nil {
        t.Fatal(err)
    }
    var walker1 walker
    
    js.Walk(&walker1, ast)
    for _, route := range walker1.Routes {
        fmt.Println(route)
    }
}
