Geoget is a tool which gives you an easy way to test the actual pre-release of PC/GEOS (https://github.com/bluewaysw/pcgeos) in combination with the Basebox Release (https://github.com/bluewaysw/pcgeos-basebox).

CAUTION: data you create with PC/GEOS Ensemble won't be preserved on an update, this is for testing / debbugging purposes only.

The easiest way to use it is to just launch geoget, e.g. geoget-linux or geoget-win64.exe from Explorer. After doing so you will find a folder called "geospc" in your home folder. You can start the launcher "ensemble.cmd" in the newly created directory.

If you want more, the advanced usage looks like this:

### How to use

```
geoget [options] [install_root]

Options:
  -f, --force            overwrite existing installation without prompt
  -g, --geos <issue>     use CI-latest-<issue> for GEOS downloads (accepts 829 or #829)
  -b, --basebox <issue>  use CI-latest-<issue> for Basebox downloads (accepts 13 or #13)
  -h, --help             show this help message
  -l, --lang <lang>      non-english GEOS language to install (only "gr" supported for now)

Arguments:
  install_root           optional install root; defaults to "geospc" under home

Defaults:
  If no issue flags are provided, CI-latest is used.
```

=====================================================================

Geoget ist ein Werkzeug, das eine einfache Möglichkeit bietet, die aktuelle Vorabversion von PC/GEOS (https://github.com/bluewaysw/pcgeos) in Kombination mit der Basebox-Version (https://github.com/bluewaysw/pcgeos-basebox) zu testen.

ACHTUNG: Daten, die Sie mit PC/GEOS Ensemble erstellen, bleiben bei einem Update nicht erhalten. Dies ist ausschließlich für Test- und Debugging-Zwecke gedacht.

### Nutzung

Der einfachste Weg ist, geoget einfach zu starten, z. B. geoget-linux oder geoget-win64.exe aus dem Explorer. Danach finden Sie in Ihrem Home-Verzeichnis einen Ordner namens "geospc". In dem neu angelegten Verzeichnis können Sie den Launcher "ensemble.cmd" starten.

Wenn Sie mehr wollen, sieht die erweiterte Nutzung so aus:

```
geoget [Optionen] [install_root]

Optionen:
  -f, --force            vorhandene Installation ohne Rückfrage überschreiben
  -g, --geos <issue>     CI-latest-<issue> für GEOS-Downloads verwenden (akzeptiert 829 oder #829)
  -b, --basebox <issue>  CI-latest-<issue> für Basebox-Downloads verwenden (akzeptiert 13 oder #13)
  -h, --help             diese Hilfe anzeigen
  -l, --lang <lang>      nicht-englische GEOS-Sprache installieren (derzeit nur "gr" unterstützt)

Argumente:
  install_root           optionales Installationsverzeichnis; Standard ist "geospc" im Home-Verzeichnis

Standardverhalten:
  Wenn keine Issue-Optionen angegeben werden, wird CI-latest verwendet.
```
