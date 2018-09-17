package handlers

import (
	//"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spf13/viper"
	"net/http"
)

func HandleHome(v *viper.Viper, w http.ResponseWriter, r *http.Request, vars map[string]string, qs railgun.QueryString, messages chan interface{}, collectionsList []railgun.Collection, collectionsByName map[string]railgun.Collection) {

	str := `
    <!doctype html>
    <html lang="en">
      <head>
        <link rel="stylesheet" href="https://cdn.rawgit.com/openlayers/openlayers.github.io/master/en/v5.2.0/css/ol.css" type="text/css">
        <style>
          .map {
            height: 100%;
            width: 100%;
          }
          #info {
            z-index: 1;
            opacity: 0;
            position: absolute;
            bottom: 0;
            left: 0;
            margin: 0;
            background: rgba(0,60,136,0.7);
            color: white;
            border: 0;
            transition: opacity 100ms ease-in;
          }
        </style>
        <script src="https://cdn.rawgit.com/openlayers/openlayers.github.io/master/en/v5.2.0/build/ol.js"></script>
        <title>Railgun</title>
      </head>
      <body>
        <div>
          <input
            id="dfl"
            type="text"
            style="width: 80%;font-size:1.4rem;line-height:1.8rem;font-family:Verdana;padding: 10px;">
          <button id="updateLayer" class="block">Button</button>
        </div>
        <div id="map" class="map">
          <pre id="info"/>
        </div>
        <script type="text/javascript">
          var input = document.getElementById("dfl");
          input.value = (new URL(window.location.href)).searchParams.get("dfl")
          input.addEventListener("keyup", function(event) {
            event.preventDefault();
            if (event.keyCode === 13) {
              document.getElementById("updateLayer").click();
            }
          });
          var btn = document.getElementById("updateLayer");
          btn.addEventListener("click", function() {
            let url = new URL(window.location.href);
            let vectorUrl = createVectorUrl(
              "dc-amenities",
              "filter(@, \""+document.getElementById("dfl").value+"\")",
              url.searchParams.get("limit"));
            let vectorSource = createVectorSource(vectorUrl);
            vectorLayer.setSource(vectorSource);
          });
        </script>
        <script>
          var createVectorUrl = function(collection, dfl, limit) {
            var vectorUrl = "/collections/"+collection+"/tiles/{z}/{x}/{y}.geojson";
            var qs = {};
            if (dfl != undefined && dfl.length > 0 ) {
              qs.dfl = encodeURI(dfl);
            }
            if (limit != undefined && limit.length > 0 ) {
              qs.limit = limit;
            }
            var keys = Object.keys(qs);
            if(keys.length > 0) {
             vectorUrl += "?";
             for(var i = 0; i < keys.length; i++) {
               if(i > 0) {
                 vectorUrl += "&"
               }
               vectorUrl += keys[i] + "=" + qs[keys[i]];
             }
            };
            return vectorUrl;
          };
          
          var createVectorSource = function(vectorUrl) {
            return new ol.source.VectorTile({
              format: new ol.format.GeoJSON({"featureProjection": "EPSG:3857", "dataProjection": "EPSG:4326"}),
              maxZoom: 19,
              url: vectorUrl,
              tileLoadFunction: function(tile, url) {
                tile.setLoader(function() {
                  var xhr = new XMLHttpRequest();
                  xhr.open('GET', url);
                  xhr.onload = function() {
                    var json = JSON.parse(xhr.responseText);
                    var format = tile.getFormat();
                    tile.setFeatures(format.readFeatures(json));
                    tile.setProjection(format.defaultFeatureProjection);
                  }
                  xhr.send();
                });
              }
            });
          };
        </script>
        <script type="text/javascript">

          var baseLayer = new ol.layer.Tile({
            source: new ol.source.OSM()
          });
          let url = new URL(window.location.href);
          var vectorUrl = createVectorUrl(
            "dc-amenities",
            "filter(@, \""+(url.searchParams.get("dfl"))+"\")",
            url.searchParams.get("limit"));
          var vectorSource = createVectorSource(vectorUrl);
          var vectorLayer = new ol.layer.VectorTile({
            source: vectorSource,
            style: function(feature) {
              return new ol.style.Style({
                image: new ol.style.Circle({
                  radius: 8,
                  fill: new ol.style.Fill({color:'rgba(0,0,255,0.8)'}),
                  stroke: new ol.style.Stroke({color: 'black', width: 2})
                })
              });
            }
          });
          var map = new ol.Map({
            target: 'map',
            layers: [baseLayer, vectorLayer],
            view: new ol.View({
              center: ol.proj.fromLonLat([0, 0]),
              zoom: 3
            })
          });
              
          map.on('pointermove', function(event) {
            let info = document.getElementById('info')
            var features = map.getFeaturesAtPixel(event.pixel);
            if (!features) {
              info.innerText = '';
              info.style.opacity = 0;
              return;
            }
            var f = features[0];
            //var properties = features[0].getProperties();
            var properties = f.getKeys()
              .filter(x => x != "geometry")
              .map(x => [x, f.get(x)])
              .reduce(function(o, x) { o[x[0]] = x[1]; return o; }, {});
            info.innerText = JSON.stringify(properties, null, 2);
            info.style.opacity = 1;
          });
          
        </script>
      </body>
    </html>
  `

	w.Write([]byte(str))

}
