(() => {
const LANGS = ["en","ru","de","fr","es","pl","lt"];
const TZ_CORE = [["UTC","UTC"],["Europe/Minsk","Europe/Minsk"],["Europe/Moscow","Europe/Moscow"],["Europe/Kyiv","Europe/Kyiv"],["Europe/Warsaw","Europe/Warsaw"],["Europe/Berlin","Europe/Berlin"],["Europe/London","Europe/London"],["America/New_York","America/New_York"],["America/Chicago","America/Chicago"],["America/Los_Angeles","America/Los_Angeles"],["Asia/Almaty","Asia/Almaty"],["Asia/Tokyo","Asia/Tokyo"]];
function tzList(localLabel, offsetLabel){
  return [["local", localLabel], ...TZ_CORE, ["+03:00", offsetLabel]];
}
const I18N = {
en: {
title: "Restore Google Takeout photos",
doc: "TakeoutRestore",
sub: "Files are read straight from the zip archives. Metadata from the JSON sidecars is written into the photos and videos. Source files are not modified.",
step1lab: "Sources", step2lab: "Destination", step3lab: "Run",
hSources: "Sources",
hintSources: "Add folders of takeout-*.zip archives, individual zip files, or already extracted folders. Duplicates across exports are removed automatically.",
dropHint: "Drop folders or zip files here, or use the buttons below.",
bAddFolder: "Add folder…", bAddZip: "Add zip…", bClearSrc: "Clear", bNext1: "Next",
hDest: "Where to save",
phDest: "D:\\Photos Restored",
bPickDest: "Browse…",
hintDest: "Album folders take no extra disk space (hard links on the same filesystem), and each photo is copied once. Free space needed is about the size of the photos.",
lblLayout: "Folder layout",
layouts: [["ym","YYYY\\YYYY-MM (2021\\2021-05)"],["y_m","YYYY\\MM (2021\\05)"],["ymd","YYYY\\YYYY-MM-DD (2021\\2021-05-24)"],["y","By year only (2021)"],["album","Albums only"]],
lblAlbums: "Album folders",
albums: [["hardlink","Hard links — no extra disk space (default)"],["copy","Separate copies (use disk space)"],["none","Do not create album folders"]],
lblTZ: "Time zone for dates from Google (they are in UTC)",
timezones: tzList("local — this computer (default)","+03:00 (fixed offset)"),
hintTZ: "local = this computer's time zone",
bBack2: "Back", bNext2: "Next",
hRun: "Restore",
hintRun: "Pause freezes the job in place; Stop ends it — Start again continues from saved progress.",
bGo: "Start restore", bPause: "Pause", bResume: "Resume", bStop: "Stop", bBack3: "Back",
lblPreview: "Show preview of recently imported photos",
hintPreview: "Thumbnails appear during restore for common image formats.",
hAdv: "Advanced options",
lblRename: "Rename files to “YYYY-MM-DD_HHMMSS_name”",
lblOverwrite: "Overwrite a date, GPS, or description that is already in the file (usually leave this off)",
lblSkip: "Do not copy edited versions (-edited)",
lblRating: "Favorites → 5-star rating",
lblKeywords: "Album names → keywords",
lblPeople: "People names → keywords",
lblXMP: "Write a .xmp sidecar next to AVI/BMP files",
lblForce: "Continue even if there is not enough disk space",
stopping: "Stopping…", stopped: "Stopped", paused: "Paused",
needSrc: "Add at least one source", needDest: "Choose a destination folder",
pickFail: "Could not open the folder dialog.",
gb: "GB", mb: "MB", drymark: "[Dry run]", phaseDry: "Dry run (nothing is written)",
ofFiles: (d,t)=>`${d} of ${t} files`,
ofFileBytes: (d,t)=>`${d} / ${t}`,
eta: s=>"about "+s+" left",
sumTitle: "Summary",
rows: {
media:"Media files found (every copy)", unique:"Unique", dups:"Duplicates skipped", incomplete:"Incomplete copies skipped",
will:"Will be copied", copied:"Copied", resumed:"Already there (resumed)",
djson:"Date from JSON", dexif:"Date from EXIF", dname:"Date from file name", dyear:"Year only", dnone:"No date",
gpsWill:"GPS to be written", gpsDid:"GPS written", albums:"Album links/copies", nojson:"Files without JSON",
unmatched:"JSON without a file", errors:"Errors", space:"Space needed / free"
},
openReport:"Open report", openLog:"Log", openDest:"Open the photos folder"
},
ru: {
title: "Восстановление фото из Google Takeout",
doc: "Восстановление Google Фото",
sub: "Файлы читаются прямо из zip-архивов, метаданные из JSON записываются в фото и видео. Исходные файлы не изменяются.",
step1lab: "Источники", step2lab: "Назначение", step3lab: "Запуск",
hSources: "Источники",
hintSources: "Добавьте папки с архивами takeout-*.zip, отдельные zip-файлы или уже распакованные папки. Дубликаты между выгрузками убираются автоматически.",
dropHint: "Перетащите папки или zip сюда либо нажмите кнопки ниже.",
bAddFolder: "Добавить папку…", bAddZip: "Добавить zip…", bClearSrc: "Очистить", bNext1: "Далее",
hDest: "Куда сохранить",
phDest: "D:\\Фото (восстановлено)",
bPickDest: "Выбрать…",
hintDest: "Папки альбомов не занимают лишнего места (жёсткие ссылки на том же диске), а сами фото копируются один раз. Нужно свободное место примерно равное объёму фото.",
lblLayout: "Структура папок",
layouts: [["ym","ГГГГ\\ГГГГ-ММ (2021\\2021-05)"],["y_m","ГГГГ\\ММ (2021\\05)"],["ymd","ГГГГ\\ГГГГ-ММ-ДД (2021\\2021-05-24)"],["y","Только по годам (2021)"],["album","Только по альбомам"]],
lblAlbums: "Папки альбомов",
albums: [["hardlink","Жёсткие ссылки — не занимают место (по умолчанию)"],["copy","Отдельные копии (займут место)"],["none","Не создавать папки альбомов"]],
lblTZ: "Часовой пояс для дат из Google (они в UTC)",
timezones: tzList("local — этот компьютер (по умолчанию)","+03:00 (фиксированный сдвиг)"),
hintTZ: "local = часовой пояс этого компьютера",
bBack2: "Назад", bNext2: "Далее",
hRun: "Восстановление",
hintRun: "Пауза останавливает текущий прогон на месте; Остановить завершает его — повторный запуск продолжит с сохранённого прогресса.",
bGo: "Начать восстановление", bPause: "Пауза", bResume: "Продолжить", bStop: "Остановить", bBack3: "Назад",
lblPreview: "Показывать предпросмотр недавно импортированных фото",
hintPreview: "Миниатюры появляются во время восстановления для обычных форматов изображений.",
hAdv: "Дополнительные параметры",
lblRename: "Переименовать файлы в «ГГГГ-ММ-ДД_ЧЧММСС_имя»",
lblOverwrite: "Перезаписывать уже имеющиеся в файле дату/GPS/описание (обычно не нужно)",
lblSkip: "Не копировать отредактированные версии (-edited / -изменено)",
lblRating: "Избранное → рейтинг 5 звёзд",
lblKeywords: "Названия альбомов → ключевые слова",
lblPeople: "Имена людей → ключевые слова",
lblXMP: "Для AVI/BMP создавать рядом файл .xmp с метаданными",
lblForce: "Продолжать даже если не хватает места на диске",
stopping: "Остановка…", stopped: "Остановлено", paused: "Пауза",
needSrc: "Добавьте хотя бы один источник", needDest: "Укажите папку назначения",
pickFail: "Не удалось открыть окно выбора папки.",
gb: "ГБ", mb: "МБ", drymark: "[Пробный прогон]", phaseDry: "Пробный прогон (ничего не записывается)",
ofFiles: (d,t)=>`${d} из ${t} файлов`,
ofFileBytes: (d,t)=>`${d} / ${t}`,
eta: s=>"осталось ≈ "+s.replace("h"," ч ").replace("m"," мин ").replace(/([\d.]+)s/,"$1 с"),
sumTitle: "Итог",
rows: {
media:"Найдено медиафайлов (всех копий)", unique:"Уникальных", dups:"Дубликатов пропущено", incomplete:"Недораспакованных копий пропущено",
will:"Будет скопировано", copied:"Скопировано", resumed:"Уже было (повторный запуск)",
djson:"Дата из JSON", dexif:"Дата из EXIF", dname:"Дата из имени файла", dyear:"Только год", dnone:"Без даты",
gpsWill:"GPS будет записан", gpsDid:"GPS записан", albums:"Ссылок/копий в альбомах", nojson:"Файлов без JSON",
unmatched:"JSON без файла", errors:"Ошибок", space:"Нужно места / свободно"
},
openReport:"Открыть отчёт", openLog:"Журнал", openDest:"Открыть папку с фото"
},
de: {
title: "Google-Takeout-Fotos wiederherstellen",
doc: "TakeoutRestore",
sub: "Dateien werden direkt aus den ZIP-Archiven gelesen. Metadaten aus den JSON-Begleitdateien werden in Fotos und Videos geschrieben. Quelldateien bleiben unverändert.",
step1lab: "Quellen", step2lab: "Ziel", step3lab: "Start",
hSources: "Quellen",
hintSources: "Ordner mit takeout-*.zip-Archiven, einzelne ZIP-Dateien oder bereits entpackte Ordner hinzufügen. Duplikate zwischen Exporten werden automatisch entfernt.",
dropHint: "Ordner oder ZIP-Dateien hier ablegen oder die Buttons unten nutzen.",
bAddFolder: "Ordner hinzufügen…", bAddZip: "ZIP hinzufügen…", bClearSrc: "Leeren", bNext1: "Weiter",
hDest: "Speicherort",
phDest: "D:\\Fotos wiederhergestellt",
bPickDest: "Durchsuchen…",
hintDest: "Albenordner brauchen keinen extra Speicherplatz (Hardlinks); jedes Foto wird einmal kopiert. Freier Speicher etwa in Foto-Größe.",
lblLayout: "Ordnerstruktur",
layouts: [["ym","YYYY\\YYYY-MM (2021\\2021-05)"],["y_m","YYYY\\MM (2021\\05)"],["ymd","YYYY\\YYYY-MM-DD (2021\\2021-05-24)"],["y","Nur nach Jahr (2021)"],["album","Nur Alben"]],
lblAlbums: "Albenordner",
albums: [["hardlink","Hardlinks — kein extra Speicherplatz (Standard)"],["copy","Separate Kopien (brauchen Speicherplatz)"],["none","Keine Albenordner anlegen"]],
lblTZ: "Zeitzone für Daten von Google (sie sind in UTC)",
timezones: tzList("local — dieser Computer (Standard)","+03:00 (fester Versatz)"),
hintTZ: "local = Zeitzone dieses Computers",
bBack2: "Zurück", bNext2: "Weiter",
hRun: "Wiederherstellen",
hintRun: "Pause friert den Lauf ein; Stopp beendet ihn — erneutes Starten setzt den Fortschritt fort.",
bGo: "Wiederherstellung starten", bPause: "Pause", bResume: "Fortsetzen", bStop: "Stopp", bBack3: "Zurück",
lblPreview: "Vorschau kürzlich importierter Fotos anzeigen",
hintPreview: "Miniaturen erscheinen während der Wiederherstellung für gängige Bildformate.",
hAdv: "Erweiterte Optionen",
lblRename: "Dateien in „YYYY-MM-DD_HHMMSS_name“ umbenennen",
lblOverwrite: "Vorhandenes Datum, GPS oder Beschreibung überschreiben (meist auslassen)",
lblSkip: "Bearbeitete Versionen nicht kopieren (-edited)",
lblRating: "Favoriten → 5-Sterne-Bewertung",
lblKeywords: "Albumnamen → Schlüsselwörter",
lblPeople: "Personennamen → Schlüsselwörter",
lblXMP: "Neben AVI/BMP eine .xmp-Sidecar schreiben",
lblForce: "Auch bei zu wenig Speicherplatz fortfahren",
stopping: "Wird gestoppt…", stopped: "Gestoppt", paused: "Pausiert",
needSrc: "Mindestens eine Quelle hinzufügen", needDest: "Zielordner wählen",
pickFail: "Ordnerdialog konnte nicht geöffnet werden.",
gb: "GB", mb: "MB", drymark: "[Probelauf]", phaseDry: "Probelauf (es wird nichts geschrieben)",
ofFiles: (d,t)=>`${d} von ${t} Dateien`,
ofFileBytes: (d,t)=>`${d} / ${t}`,
eta: s=>"noch ca. "+s,
sumTitle: "Zusammenfassung",
rows: {
media:"Mediendateien gefunden (alle Kopien)", unique:"Einzigartig", dups:"Duplikate übersprungen", incomplete:"Unvollständige Kopien übersprungen",
will:"Wird kopiert", copied:"Kopiert", resumed:"Bereits vorhanden (fortgesetzt)",
djson:"Datum aus JSON", dexif:"Datum aus EXIF", dname:"Datum aus Dateiname", dyear:"Nur Jahr", dnone:"Kein Datum",
gpsWill:"GPS wird geschrieben", gpsDid:"GPS geschrieben", albums:"Album-Links/Kopien", nojson:"Dateien ohne JSON",
unmatched:"JSON ohne Datei", errors:"Fehler", space:"Benötigt / frei"
},
openReport:"Bericht öffnen", openLog:"Protokoll", openDest:"Fotoordner öffnen"
},
fr: {
title: "Restaurer les photos Google Takeout",
doc: "TakeoutRestore",
sub: "Les fichiers sont lus directement depuis les archives zip. Les métadonnées des sidecars JSON sont écrites dans les photos et vidéos. Les fichiers sources ne sont pas modifiés.",
step1lab: "Sources", step2lab: "Destination", step3lab: "Lancer",
hSources: "Sources",
hintSources: "Ajoutez des dossiers d’archives takeout-*.zip, des fichiers zip ou des dossiers déjà extraits. Les doublons entre exports sont retirés automatiquement.",
dropHint: "Déposez des dossiers ou des zip ici, ou utilisez les boutons ci-dessous.",
bAddFolder: "Ajouter un dossier…", bAddZip: "Ajouter un zip…", bClearSrc: "Effacer", bNext1: "Suivant",
hDest: "Emplacement d’enregistrement",
phDest: "D:\\Photos restaurées",
bPickDest: "Parcourir…",
hintDest: "Les dossiers d’albums n’utilisent pas d’espace supplémentaire (liens physiques) ; chaque photo est copiée une fois. Espace libre ≈ taille des photos.",
lblLayout: "Structure des dossiers",
layouts: [["ym","YYYY\\YYYY-MM (2021\\2021-05)"],["y_m","YYYY\\MM (2021\\05)"],["ymd","YYYY\\YYYY-MM-DD (2021\\2021-05-24)"],["y","Par année seulement (2021)"],["album","Albums seulement"]],
lblAlbums: "Dossiers d’albums",
albums: [["hardlink","Liens physiques — pas d’espace en plus (défaut)"],["copy","Copies séparées (utilisent de l’espace)"],["none","Ne pas créer de dossiers d’albums"]],
lblTZ: "Fuseau horaire des dates Google (elles sont en UTC)",
timezones: tzList("local — cet ordinateur (défaut)","+03:00 (décalage fixe)"),
hintTZ: "local = fuseau horaire de cet ordinateur",
bBack2: "Retour", bNext2: "Suivant",
hRun: "Restaurer",
hintRun: "Pause fige le travail ; Arrêter le termine — un nouveau démarrage reprend la progression.",
bGo: "Démarrer la restauration", bPause: "Pause", bResume: "Reprendre", bStop: "Arrêter", bBack3: "Retour",
lblPreview: "Afficher l’aperçu des photos récemment importées",
hintPreview: "Les miniatures apparaissent pendant la restauration, pour les formats image courants.",
hAdv: "Options avancées",
lblRename: "Renommer en « YYYY-MM-DD_HHMMSS_nom »",
lblOverwrite: "Écraser une date, un GPS ou une description déjà présents (souvent à laisser désactivé)",
lblSkip: "Ne pas copier les versions modifiées (-edited)",
lblRating: "Favoris → note 5 étoiles",
lblKeywords: "Noms d’albums → mots-clés",
lblPeople: "Noms de personnes → mots-clés",
lblXMP: "Écrire un sidecar .xmp à côté des fichiers AVI/BMP",
lblForce: "Continuer même s’il n’y a pas assez d’espace disque",
stopping: "Arrêt…", stopped: "Arrêté", paused: "En pause",
needSrc: "Ajoutez au moins une source", needDest: "Choisissez un dossier de destination",
pickFail: "Impossible d’ouvrir le dialogue de dossier.",
gb: "Go", mb: "Mo", drymark: "[Essai à blanc]", phaseDry: "Essai à blanc (rien n’est écrit)",
ofFiles: (d,t)=>`${d} sur ${t} fichiers`,
ofFileBytes: (d,t)=>`${d} / ${t}`,
eta: s=>"environ "+s+" restant",
sumTitle: "Résumé",
rows: {
media:"Fichiers multimédias trouvés (toutes copies)", unique:"Uniques", dups:"Doublons ignorés", incomplete:"Copies incomplètes ignorées",
will:"Seront copiés", copied:"Copiés", resumed:"Déjà présents (reprise)",
djson:"Date depuis JSON", dexif:"Date depuis EXIF", dname:"Date depuis le nom", dyear:"Année seule", dnone:"Sans date",
gpsWill:"GPS à écrire", gpsDid:"GPS écrit", albums:"Liens/copies d’albums", nojson:"Fichiers sans JSON",
unmatched:"JSON sans fichier", errors:"Erreurs", space:"Espace nécessaire / libre"
},
openReport:"Ouvrir le rapport", openLog:"Journal", openDest:"Ouvrir le dossier photos"
},
es: {
title: "Restaurar fotos de Google Takeout",
doc: "TakeoutRestore",
sub: "Los archivos se leen directamente de los zip. Los metadatos de los sidecars JSON se escriben en fotos y vídeos. Los archivos de origen no se modifican.",
step1lab: "Fuentes", step2lab: "Destino", step3lab: "Ejecutar",
hSources: "Fuentes",
hintSources: "Añada carpetas con takeout-*.zip, archivos zip o carpetas ya extraídas. Los duplicados entre exportaciones se eliminan automáticamente.",
dropHint: "Suelte carpetas o zip aquí, o use los botones de abajo.",
bAddFolder: "Añadir carpeta…", bAddZip: "Añadir zip…", bClearSrc: "Borrar", bNext1: "Siguiente",
hDest: "Dónde guardar",
phDest: "D:\\Fotos restauradas",
bPickDest: "Examinar…",
hintDest: "Las carpetas de álbumes no ocupan espacio extra (enlaces duros); cada foto se copia una vez. Espacio libre ≈ tamaño de las fotos.",
lblLayout: "Estructura de carpetas",
layouts: [["ym","YYYY\\YYYY-MM (2021\\2021-05)"],["y_m","YYYY\\MM (2021\\05)"],["ymd","YYYY\\YYYY-MM-DD (2021\\2021-05-24)"],["y","Solo por año (2021)"],["album","Solo álbumes"]],
lblAlbums: "Carpetas de álbumes",
albums: [["hardlink","Enlaces duros — sin espacio extra (predeterminado)"],["copy","Copias separadas (usan espacio)"],["none","No crear carpetas de álbumes"]],
lblTZ: "Zona horaria de las fechas de Google (están en UTC)",
timezones: tzList("local — este equipo (predeterminado)","+03:00 (desplazamiento fijo)"),
hintTZ: "local = zona horaria de este equipo",
bBack2: "Atrás", bNext2: "Siguiente",
hRun: "Restaurar",
hintRun: "Pausa congela el trabajo; Detener lo finaliza — al iniciar de nuevo se continúa desde el progreso guardado.",
bGo: "Iniciar restauración", bPause: "Pausa", bResume: "Reanudar", bStop: "Detener", bBack3: "Atrás",
lblPreview: "Mostrar vista previa de fotos recién importadas",
hintPreview: "Las miniaturas aparecen durante la restauración para formatos de imagen comunes.",
hAdv: "Opciones avanzadas",
lblRename: "Renombrar a «YYYY-MM-DD_HHMMSS_nombre»",
lblOverwrite: "Sobrescribir fecha, GPS o descripción ya presentes (suele dejarse desactivado)",
lblSkip: "No copiar versiones editadas (-edited)",
lblRating: "Favoritos → valoración de 5 estrellas",
lblKeywords: "Nombres de álbum → palabras clave",
lblPeople: "Nombres de personas → palabras clave",
lblXMP: "Escribir un sidecar .xmp junto a AVI/BMP",
lblForce: "Continuar aunque no haya suficiente espacio en disco",
stopping: "Deteniendo…", stopped: "Detenido", paused: "En pausa",
needSrc: "Añada al menos una fuente", needDest: "Elija una carpeta de destino",
pickFail: "No se pudo abrir el diálogo de carpeta.",
gb: "GB", mb: "MB", drymark: "[Simulación]", phaseDry: "Simulación (no se escribe nada)",
ofFiles: (d,t)=>`${d} de ${t} archivos`,
ofFileBytes: (d,t)=>`${d} / ${t}`,
eta: s=>"quedan unos "+s,
sumTitle: "Resumen",
rows: {
media:"Archivos multimedia encontrados (todas las copias)", unique:"Únicos", dups:"Duplicados omitidos", incomplete:"Copias incompletas omitidas",
will:"Se copiarán", copied:"Copiados", resumed:"Ya existían (reanudado)",
djson:"Fecha desde JSON", dexif:"Fecha desde EXIF", dname:"Fecha desde el nombre", dyear:"Solo año", dnone:"Sin fecha",
gpsWill:"GPS a escribir", gpsDid:"GPS escrito", albums:"Enlaces/copias de álbumes", nojson:"Archivos sin JSON",
unmatched:"JSON sin archivo", errors:"Errores", space:"Espacio necesario / libre"
},
openReport:"Abrir informe", openLog:"Registro", openDest:"Abrir carpeta de fotos"
},
pl: {
title: "Przywracanie zdjęć z Google Takeout",
doc: "TakeoutRestore",
sub: "Pliki są czytane bezpośrednio z archiwów zip. Metadane z plików JSON są zapisywane w zdjęciach i filmach. Pliki źródłowe nie są zmieniane.",
step1lab: "Źródła", step2lab: "Cel", step3lab: "Uruchom",
hSources: "Źródła",
hintSources: "Dodaj foldery z archiwami takeout-*.zip, pojedyncze pliki zip lub już rozpakowane foldery. Duplikaty między eksportami są usuwane automatycznie.",
dropHint: "Upuść foldery lub zip tutaj albo użyj przycisków poniżej.",
bAddFolder: "Dodaj folder…", bAddZip: "Dodaj zip…", bClearSrc: "Wyczyść", bNext1: "Dalej",
hDest: "Gdzie zapisać",
phDest: "D:\\Zdjęcia przywrócone",
bPickDest: "Przeglądaj…",
hintDest: "Foldery albumów nie zajmują dodatkowego miejsca (twarde łącza); każde zdjęcie kopiowane jest raz. Potrzeba wolnego miejsca ≈ rozmiar zdjęć.",
lblLayout: "Układ folderów",
layouts: [["ym","YYYY\\YYYY-MM (2021\\2021-05)"],["y_m","YYYY\\MM (2021\\05)"],["ymd","YYYY\\YYYY-MM-DD (2021\\2021-05-24)"],["y","Tylko według roku (2021)"],["album","Tylko albumy"]],
lblAlbums: "Foldery albumów",
albums: [["hardlink","Twarde łącza — bez dodatkowego miejsca (domyślnie)"],["copy","Osobne kopie (zajmują miejsce)"],["none","Nie twórz folderów albumów"]],
lblTZ: "Strefa czasowa dat z Google (są w UTC)",
timezones: tzList("local — ten komputer (domyślnie)","+03:00 (stałe przesunięcie)"),
hintTZ: "local = strefa czasowa tego komputera",
bBack2: "Wstecz", bNext2: "Dalej",
hRun: "Przywracanie",
hintRun: "Pauza zamraża zadanie; Stop je kończy — ponowne uruchomienie kontynuuje z zapisanego postępu.",
bGo: "Rozpocznij przywracanie", bPause: "Pauza", bResume: "Wznów", bStop: "Stop", bBack3: "Wstecz",
lblPreview: "Pokazuj podgląd niedawno zaimportowanych zdjęć",
hintPreview: "Miniatury pojawiają się podczas przywracania dla typowych formatów obrazów.",
hAdv: "Opcje zaawansowane",
lblRename: "Zmień nazwy na „YYYY-MM-DD_HHMMSS_nazwa”",
lblOverwrite: "Nadpisz datę, GPS lub opis już obecne w pliku (zwykle wyłączone)",
lblSkip: "Nie kopiuj wersji edytowanych (-edited)",
lblRating: "Ulubione → ocena 5 gwiazdek",
lblKeywords: "Nazwy albumów → słowa kluczowe",
lblPeople: "Imiona osób → słowa kluczowe",
lblXMP: "Zapisz sidecar .xmp obok plików AVI/BMP",
lblForce: "Kontynuuj mimo braku miejsca na dysku",
stopping: "Zatrzymywanie…", stopped: "Zatrzymano", paused: "Wstrzymano",
needSrc: "Dodaj co najmniej jedno źródło", needDest: "Wybierz folder docelowy",
pickFail: "Nie udało się otworzyć okna wyboru folderu.",
gb: "GB", mb: "MB", drymark: "[Symulacja]", phaseDry: "Symulacja (nic nie jest zapisywane)",
ofFiles: (d,t)=>`${d} z ${t} plików`,
ofFileBytes: (d,t)=>`${d} / ${t}`,
eta: s=>"pozostało ok. "+s,
sumTitle: "Podsumowanie",
rows: {
media:"Znalezione pliki mediów (wszystkie kopie)", unique:"Unikalne", dups:"Pominięte duplikaty", incomplete:"Pominięte niekompletne kopie",
will:"Zostanie skopiowane", copied:"Skopiowano", resumed:"Już było (wznowienie)",
djson:"Data z JSON", dexif:"Data z EXIF", dname:"Data z nazwy pliku", dyear:"Tylko rok", dnone:"Bez daty",
gpsWill:"GPS do zapisania", gpsDid:"GPS zapisany", albums:"Łącza/kopie albumów", nojson:"Pliki bez JSON",
unmatched:"JSON bez pliku", errors:"Błędy", space:"Potrzebne / wolne miejsce"
},
openReport:"Otwórz raport", openLog:"Dziennik", openDest:"Otwórz folder ze zdjęciami"
},
lt: {
title: "Atkurti Google Takeout nuotraukas",
doc: "TakeoutRestore",
sub: "Failai skaitomi tiesiai iš zip archyvų. Metaduomenys iš JSON šalia esančių failų įrašomi į nuotraukas ir vaizdo įrašus. Šaltinio failai nekeičiami.",
step1lab: "Šaltiniai", step2lab: "Paskirtis", step3lab: "Paleisti",
hSources: "Šaltiniai",
hintSources: "Pridėkite aplankus su takeout-*.zip, atskirus zip failus arba jau išpakuotus aplankus. Dublikatai tarp eksportų pašalinami automatiškai.",
dropHint: "Numeskite aplankus ar zip čia arba naudokite mygtukus žemiau.",
bAddFolder: "Pridėti aplanką…", bAddZip: "Pridėti zip…", bClearSrc: "Išvalyti", bNext1: "Toliau",
hDest: "Kur išsaugoti",
phDest: "D:\\Nuotraukos atkurtos",
bPickDest: "Naršyti…",
hintDest: "Albumų aplankai neužima papildomos vietos (kietosios nuorodos); kiekviena nuotrauka kopijuojama vieną kartą. Reikia laisvos vietos ≈ nuotraukų dydžiui.",
lblLayout: "Aplankų struktūra",
layouts: [["ym","YYYY\\YYYY-MM (2021\\2021-05)"],["y_m","YYYY\\MM (2021\\05)"],["ymd","YYYY\\YYYY-MM-DD (2021\\2021-05-24)"],["y","Tik pagal metus (2021)"],["album","Tik albumai"]],
lblAlbums: "Albumų aplankai",
albums: [["hardlink","Kietosios nuorodos — be papildomos vietos (numatyta)"],["copy","Atskiros kopijos (užima vietą)"],["none","Nekurti albumų aplankų"]],
lblTZ: "Google datų laiko juosta (jos UTC)",
timezones: tzList("local — šis kompiuteris (numatyta)","+03:00 (fiksuotas poslinkis)"),
hintTZ: "local = šio kompiuterio laiko juosta",
bBack2: "Atgal", bNext2: "Toliau",
hRun: "Atkurti",
hintRun: "Pristabdymas sustabdo darbą; Stabdyti jį baigia — paleidus iš naujo tęsiama nuo išsaugotos pažangos.",
bGo: "Pradėti atkūrimą", bPause: "Pristabdyti", bResume: "Tęsti", bStop: "Stabdyti", bBack3: "Atgal",
lblPreview: "Rodyti neseniai importuotų nuotraukų peržiūrą",
hintPreview: "Miniatiūros atsiranda atkūrimo metu įprastiems vaizdų formatams.",
hAdv: "Išplėstinės parinktys",
lblRename: "Pervadinti į „YYYY-MM-DD_HHMMSS_vardas“",
lblOverwrite: "Perrašyti faile jau esančią datą, GPS ar aprašymą (paprastai palikite išjungta)",
lblSkip: "Nekopijuoti redaguotų versijų (-edited)",
lblRating: "Mėgstamiausi → 5 žvaigždučių įvertinimas",
lblKeywords: "Albumų pavadinimai → raktažodžiai",
lblPeople: "Žmonių vardai → raktažodžiai",
lblXMP: "Šalia AVI/BMP rašyti .xmp failą",
lblForce: "Tęsti net jei diske trūksta vietos",
stopping: "Stabdoma…", stopped: "Sustabdyta", paused: "Pristabdyta",
needSrc: "Pridėkite bent vieną šaltinį", needDest: "Pasirinkite paskirties aplanką",
pickFail: "Nepavyko atidaryti aplanko dialogo.",
gb: "GB", mb: "MB", drymark: "[Bandymas]", phaseDry: "Bandomasis paleidimas (nieko nerašoma)",
ofFiles: (d,t)=>`${d} iš ${t} failų`,
ofFileBytes: (d,t)=>`${d} / ${t}`,
eta: s=>"liko ≈ "+s,
sumTitle: "Santrauka",
rows: {
media:"Rasta medijos failų (visos kopijos)", unique:"Unikalūs", dups:"Praleisti dublikatai", incomplete:"Praleistos nepilnos kopijos",
will:"Bus nukopijuota", copied:"Nukopijuota", resumed:"Jau buvo (tęsta)",
djson:"Data iš JSON", dexif:"Data iš EXIF", dname:"Data iš failo vardo", dyear:"Tik metai", dnone:"Be datos",
gpsWill:"GPS bus įrašytas", gpsDid:"GPS įrašytas", albums:"Albumų nuorodos/kopijos", nojson:"Failai be JSON",
unmatched:"JSON be failo", errors:"Klaidos", space:"Reikia vietos / laisva"
},
openReport:"Atidaryti ataskaitą", openLog:"Žurnalas", openDest:"Atidaryti nuotraukų aplanką"
}
};

const $ = id => document.getElementById(id);
const flags = ["rename","overwrite","skipEdited","rating","albumKeywords","people","xmpSidecar","force"];
const textIds = [
  "title","subtitle","step1lab","step2lab","step3lab","hSources","hintSources","dropHint",
  "bAddFolder","bAddZip","bClearSrc","bNext1","hDest","bPickDest","hintDest","lblLayout","lblAlbums",
  "lblTZ","hintTZ","bBack2","bNext2","hRun","hintRun","bGo","bPause","bStop","bBack3",
  "lblPreview","hintPreview","hAdv",
  "lblRename","lblOverwrite","lblSkip","lblRating","lblKeywords","lblPeople","lblXMP","lblForce"
];

let lang = "en";
let sources = [];
let step = 1;
let last = null;
let stopping = false;
let api = null;

function isLang(l){ return LANGS.indexOf(l) >= 0; }

function backend() {
  if (api) return api;
  const g = window.go && window.go.main && window.go.main.DesktopApp;
  if (!g) throw new Error("DesktopApp bindings unavailable");
  api = g;
  return api;
}

function t(k){ return (I18N[lang]||I18N.en)[k]; }
function esc(s){ return String(s).replace(/[&<>]/g,c=>({"&":"&amp;","<":"&lt;",">":"&gt;"}[c])); }
function gbf(b){ return (b/1073741824).toFixed(2)+" "+t("gb"); }
function sizef(b){ if (b >= 1073741824) return gbf(b); return (b/1048576).toFixed(1)+" "+t("mb"); }

function fillSelect(id, options){
  const el = $(id), keep = el.value || el.dataset.keep || "";
  el.innerHTML = options.map(o=>`<option value="${o[0]}">${o[1]}</option>`).join("");
  if (keep) el.value = keep;
}

function remember(extra){
  let s = {};
  try { s = JSON.parse(localStorage.getItem("tr")||"{}"); } catch(e) {}
  s.lang = lang;
  s.sources = sources;
  s.dest = $("dest").value;
  s.layout = $("layout").value;
  s.albums = $("albums").value;
  s.tz = $("tz").value;
  s.previewOn = $("previewOn").checked;
  if (extra) Object.assign(s, extra);
  localStorage.setItem("tr", JSON.stringify(s));
}

function applyLang(){
  const d = I18N[lang]||I18N.en;
  document.documentElement.lang = lang;
  document.title = d.doc;
  textIds.forEach(id => { if ($(id) && d[id] != null) $(id).textContent = d[id]; });
  $("dest").placeholder = d.phDest;
  fillSelect("layout", d.layouts);
  fillSelect("albums", d.albums);
  fillSelect("tz", d.timezones);
  $("langSel").value = lang;
  if (last) render(last);
}

function setLang(l){
  lang = isLang(l) ? l : "en";
  applyLang();
  remember();
}

function setStep(n){
  step = n;
  for (let i=1;i<=3;i++){
    $("panel"+i).hidden = i!==n;
    const btn = document.querySelector(`.step[data-step="${i}"]`);
    btn.classList.toggle("on", i===n);
    btn.classList.toggle("done", i<n);
  }
}

function renderSources(){
  const ul = $("srcList");
  ul.innerHTML = sources.map((p,i)=>`<li><span title="${esc(p)}">${esc(p)}</span><button type="button" data-i="${i}" aria-label="remove">×</button></li>`).join("");
  $("bClearSrc").disabled = !sources.length;
  $("dropHint").style.display = sources.length ? "none" : "";
}

function addSource(p){
  p = String(p||"").trim().replace(/^"|"$/g,"");
  if (!p) return;
  if (!sources.includes(p)) sources.push(p);
  renderSources();
  remember();
}

function removeSource(i){
  sources.splice(i,1);
  renderSources();
  remember();
}

async function pickFolder(target){
  try {
    const r = await backend().PickFolderDialog(lang);
    if (!r || !r.path) {
      if (r && r.error) alert(t("pickFail")+"\n"+r.error);
      return;
    }
    if (target==="src") addSource(r.path);
    else $("dest").value = r.path;
    remember();
  } catch (e) {
    alert(t("pickFail")+"\n"+e);
  }
}

async function pickZips(){
  try {
    const r = await backend().PickZipDialog(lang);
    if (r && r.error) { alert(t("pickFail")+"\n"+r.error); return; }
    (r.paths||[]).forEach(addSource);
  } catch (e) {
    alert(t("pickFail")+"\n"+e);
  }
}

function collectOptions(dry){
  const o = {
    sources: sources.slice(),
    dest: $("dest").value.trim().replace(/^"|"$/g,""),
    layout: $("layout").value,
    albums: $("albums").value,
    tz: $("tz").value.trim() || "local",
    dryRun: !!dry,
    lang
  };
  flags.forEach(f => o[f] = $(f).checked);
  return o;
}

async function start(dry){
  stopping = false;
  const o = collectOptions(dry);
  if (!o.sources.length){ alert(t("needSrc")); setStep(1); return; }
  if (!o.dest){ alert(t("needDest")); setStep(2); return; }
  remember();
  const r = await backend().Start(o);
  if (r && r.error){ alert(r.error); return; }
  const s = await backend().Status();
  last = s;
  render(s);
}

function stopRun(){
  stopping = true;
  $("bStop").disabled = true;
  if ($("bPause")) $("bPause").disabled = true;
  backend().Cancel();
  if (last) render(Object.assign({}, last, {running: true, paused: false, phase: t("stopping")}));
  else render({running: true, paused: false, phase: t("stopping")});
}

function pauseRun(){
  if (!last || !last.running || last.paused) return;
  backend().Pause();
  last = Object.assign({}, last, {paused: true, phase: t("paused")});
  render(last);
}

function resumeRun(){
  if (!last || !last.running || !last.paused) return;
  backend().Resume();
  last = Object.assign({}, last, {paused: false});
  render(last);
}

function hasProgress(s){
  if (!s) return false;
  if (stopping || s.running || s.finished) return true;
  if (s.error || s.warning) return true;
  if (s.stats || s.report || s.log) return true;
  if (s.phase) return true;
  return false;
}

function renderPreview(s){
  const box = $("preview");
  const on = $("previewOn").checked;
  const items = (s && s.recent) || [];
  if (!on || !items.length) {
    box.hidden = true;
    box.innerHTML = "";
    box.dataset.sig = "";
    return;
  }
  const sig = items.map(it => it.path).join("|");
  if (box.dataset.sig === sig) {
    box.hidden = false;
    return;
  }
  box.dataset.sig = sig;
  box.hidden = false;
  box.innerHTML = items.map(it => {
    const src = "/preview?p=" + encodeURIComponent(it.path);
    return `<figure title="${esc(it.name)}"><img src="${src}" alt="" loading="lazy"><figcaption>${esc(it.name)}</figcaption></figure>`;
  }).join("");
}

function render(s){
  renderPreview(s);
  if (!hasProgress(s)) {
    $("pc").hidden = true;
    $("bGo").disabled = false;
    $("bStop").disabled = true;
    if ($("bPause")) { $("bPause").disabled = true; $("bPause").textContent = t("bPause"); }
    return;
  }
  $("pc").hidden = false;
  let phase = s.phase || "";
  if (stopping && s.running) phase = t("stopping");
  else if (s.paused && s.running) phase = t("paused");
  if (!s.running) stopping = false;
  if (s.dryRun && phase && phase !== t("phaseDry") && phase !== t("stopping") && phase !== t("stopped") && phase !== t("paused"))
    phase = t("drymark")+" "+phase;
  $("phase").textContent = phase || "—";
  const pct = s.bytesTotal ? 100*s.bytesDone/s.bytesTotal : (s.total ? 100*s.done/s.total : 0);
  $("barf").style.width = pct.toFixed(1)+"%";
  $("cnt").textContent = s.total ? t("ofFiles")(s.done,s.total)+(s.bytesTotal?` · ${gbf(s.bytesDone)} / ${gbf(s.bytesTotal)}`:"") : "";
  $("eta").textContent = s.running && s.eta ? t("eta")(s.eta) : "";
  let cur = "";
  if (s.running) {
    cur = s.current || "";
    if (s.fileSize > 0) cur += (cur ? " · " : "") + t("ofFileBytes")(sizef(s.fileBytes||0), sizef(s.fileSize));
    if (s.currentNote) cur += (cur ? " — " : "") + s.currentNote;
  }
  $("cur").textContent = cur;
  let m = "";
  if (s.warning) m += `<div class="warn">⚠ ${esc(s.warning)}</div>`;
  if (s.error) m += `<div class="err">${esc(s.error)}</div>`;
  $("msg").innerHTML = m;
  $("bGo").disabled = !!s.running;
  $("bStop").disabled = !s.running || !!stopping;
  if ($("bPause")) {
    $("bPause").disabled = !s.running || !!stopping;
    $("bPause").textContent = s.paused ? t("bResume") : t("bPause");
  }
  if (s.finished && s.stats){
    const st = s.stats, rows = t("rows");
    const data = [
      [rows.media, st.MediaEntries],
      [rows.unique, st.Unique+" ("+gbf(st.UniqueBytes)+")"],
      [rows.dups, st.Duplicates],
      [rows.incomplete, st.Incomplete],
      [s.dryRun?rows.will:rows.copied, st.Copied],
      [rows.resumed, st.Resumed],
      [rows.djson, st.DateJSON],
      [rows.dexif, st.DateEXIF],
      [rows.dname, st.DateName],
      [rows.dyear, st.DateYear],
      [rows.dnone, st.DateNone],
      [s.dryRun?rows.gpsWill:rows.gpsDid, st.GPSWritten],
      [rows.albums, (st.AlbumLinks||0)+(st.AlbumCopies||0)],
      [rows.nojson, st.MediaNoJSON],
      [rows.unmatched, st.JSONUnmatched],
      [rows.errors, st.Errors],
      [rows.space, gbf(st.NeedBytes)+" / "+gbf(st.FreeBytes)]
    ];
    $("sum").innerHTML = `<h2 style="margin-top:14px">${esc(t("sumTitle"))}</h2><table>${
      data.map(r=>`<tr><td>${esc(r[0])}</td><td class="n">${esc(r[1])}</td></tr>`).join("")
    }</table><div class="links">${
      (s.report?`<a id="aReport">${esc(t("openReport"))}</a>`:"")+
      (s.log?`<a id="aLog">${esc(t("openLog"))}</a>`:"")+
      (!s.dryRun && s.dest?`<a id="aDest">${esc(t("openDest"))}</a>`:"")
    }</div>`;
    const ar = $("aReport"); if (ar) ar.onclick = () => backend().OpenPath("report");
    const al = $("aLog"); if (al) al.onclick = () => backend().OpenPath("log");
    const ad = $("aDest"); if (ad) ad.onclick = () => backend().OpenPath("dest");
  } else {
    $("sum").innerHTML = "";
  }
}

function wireWindowControls(){
  const rt = window.runtime;
  if (!rt) return;
  $("winMin").onclick = () => rt.WindowMinimise && rt.WindowMinimise();
  $("winMax").onclick = () => rt.WindowToggleMaximise && rt.WindowToggleMaximise();
  $("winClose").onclick = () => rt.Quit && rt.Quit();
  // Double-click title bar to maximize, like a normal window.
  $("titlebar").addEventListener("dblclick", e => {
    if (e.target.closest(".titlebar-controls")) return;
    if (rt.WindowToggleMaximise) rt.WindowToggleMaximise();
  });
}

function wire(){
  wireWindowControls();
  $("langSel").onchange = () => setLang($("langSel").value);
  document.querySelectorAll(".step").forEach(btn => {
    btn.onclick = () => setStep(Number(btn.dataset.step));
  });
  $("bNext1").onclick = () => {
    if (!sources.length){ alert(t("needSrc")); return; }
    setStep(2);
  };
  $("bBack2").onclick = () => setStep(1);
  $("bNext2").onclick = () => {
    if (!$("dest").value.trim()){ alert(t("needDest")); return; }
    remember();
    setStep(3);
  };
  $("bBack3").onclick = () => setStep(2);
  $("bAddFolder").onclick = () => pickFolder("src");
  $("bAddZip").onclick = () => pickZips();
  $("bClearSrc").onclick = () => { sources = []; renderSources(); remember(); };
  $("bPickDest").onclick = () => pickFolder("dest");
  $("bGo").onclick = () => start(false);
  $("bStop").onclick = () => stopRun();
  if ($("bPause")) $("bPause").onclick = () => { if (last && last.paused) resumeRun(); else pauseRun(); };
  $("previewOn").onchange = () => {
    remember();
    renderPreview(last || {});
  };
  $("srcList").onclick = e => {
    const b = e.target.closest("button[data-i]");
    if (b) removeSource(Number(b.dataset.i));
  };
  ["dest","layout","albums","tz"].forEach(id => $(id).addEventListener("change", () => remember()));

  const drop = $("drop");
  drop.addEventListener("dragover", e => { e.preventDefault(); drop.classList.add("over"); });
  drop.addEventListener("dragleave", () => drop.classList.remove("over"));
  drop.addEventListener("drop", e => {
    e.preventDefault();
    drop.classList.remove("over");
    const files = [...(e.dataTransfer.files||[])];
    files.forEach(f => {
      // WebView may expose path on Chromium builds
      const p = f.path || f.name;
      if (p && (p.includes("\\") || p.includes("/") || /\.zip$/i.test(p))) addSource(p);
    });
  });
}

async function init(){
  try {
    const s = JSON.parse(localStorage.getItem("tr")||"{}");
    if (Array.isArray(s.sources)) sources = s.sources.filter(Boolean);
    if (s.dest) $("dest").value = s.dest;
    if (s.layout) $("layout").dataset.keep = s.layout;
    if (s.albums) $("albums").dataset.keep = s.albums;
    if (s.tz) $("tz").dataset.keep = s.tz;
    if (isLang(s.lang)) lang = s.lang;
    if (typeof s.previewOn === "boolean") $("previewOn").checked = s.previewOn;
  } catch(e) {}

  wire();
  applyLang();
  renderSources();
  setStep(1);

  try {
    const cfg = await backend().Config();
    if (cfg) {
      if (cfg.langExplicit && isLang(cfg.defaultLang)) lang = cfg.defaultLang;
      else if (!localStorage.getItem("tr") && isLang(cfg.defaultLang)) lang = cfg.defaultLang;
      applyLang();
    }
  } catch(e) {
    console.warn(e);
  }

  if (window.runtime && window.runtime.EventsOn) {
    window.runtime.EventsOn("progress", s => { last = s; render(s); });
  } else {
    setInterval(async () => {
      try {
        const s = await backend().Status();
        last = s;
        render(s);
      } catch(e) {}
    }, 700);
  }

  try {
    const s = await backend().Status();
    if (s && (s.running || s.finished)) { last = s; render(s); }
  } catch(e) {}
}

document.addEventListener("DOMContentLoaded", init);
})();
