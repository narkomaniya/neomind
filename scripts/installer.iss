#define MyAppName "NeoMind"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "narkomaniya"
#define MyAppURL "https://github.com/narkomaniya/neomind"
#define MyAppExeName "neomind-gui.exe"

[Setup]
AppId={{B7E3B1B0-5B1F-4C0E-9E4B-8E7A2C3D6F10}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}/releases
DefaultDirName={localappdata}\NeoMind
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
DisableWelcomePage=no
PrivilegesRequired=lowest
OutputDir=..\dist\installer
OutputBaseFilename=NeoMind-Setup-{#MyAppVersion}
SetupIconFile=..\assets\neomind.ico
UninstallDisplayIcon={app}\{#MyAppExeName}
UninstallDisplayName={#MyAppName}
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern
WizardImageFile=installer\wizard.bmp
WizardSmallImageFile=installer\wizard-small.bmp
WizardImageStretch=no
ShowLanguageDialog=yes
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
MinVersion=10.0.17763
CloseApplications=yes
VersionInfoVersion=1.0.0.0
VersionInfoCompany={#MyAppPublisher}
VersionInfoDescription={#MyAppName} installer
VersionInfoProductName={#MyAppName}
VersionInfoProductVersion={#MyAppVersion}

[Languages]
Name: "en"; MessagesFile: "compiler:Default.isl"
Name: "ru"; MessagesFile: "compiler:Languages\Russian.isl"
Name: "uk"; MessagesFile: "compiler:Languages\Ukrainian.isl"
Name: "es"; MessagesFile: "compiler:Languages\Spanish.isl"
Name: "de"; MessagesFile: "compiler:Languages\German.isl"
Name: "fr"; MessagesFile: "compiler:Languages\French.isl"
Name: "zh"; MessagesFile: "compiler:Languages\ChineseSimplified.isl"
Name: "ja"; MessagesFile: "compiler:Languages\Japanese.isl"

[CustomMessages]
en.WelcomeTitle=Welcome to NeoMind
en.WelcomeText=Your notes and code in one fast local workspace. Choose a language and the starter packs you want - everything stays on this PC.
en.TypeFull=Full installation (app and all starter packs)
en.TypeCompact=Compact (app only)
en.TypeCustom=Custom selection
en.CompCore=NeoMind application (required)
en.CompWeb=Web development - HTML, CSS and JavaScript starter files
en.CompGame=Game development - a small canvas game loop
en.CompCyber=Cybersecurity - a safe recon script and notes template
en.CompData=Data and AI - a dependency-free CSV analysis starter
en.CompEditor=Editor and productivity - snippets and a sample keymap
en.AddToPath=Add NeoMind to the user PATH (CLI)

ru.WelcomeTitle=Добро пожаловать в NeoMind
ru.WelcomeText=Заметки и код в одном быстром локальном пространстве. Выберите язык и наборы — всё остаётся на этом компьютере.
ru.TypeFull=Полная установка (приложение и все наборы)
ru.TypeCompact=Компактная (только приложение)
ru.TypeCustom=Выборочная установка
ru.CompCore=Приложение NeoMind (обязательно)
ru.CompWeb=Веб-разработка — стартовые файлы HTML, CSS и JavaScript
ru.CompGame=Геймдев — небольшой игровой цикл на canvas
ru.CompCyber=Кибербезопасность — безопасный скрипт разведки и шаблон заметок
ru.CompData=Данные и ИИ — анализ CSV без зависимостей
ru.CompEditor=Редактор и продуктивность — сниппеты и пример раскладки
ru.AddToPath=Добавить NeoMind в PATH пользователя (CLI)

uk.WelcomeTitle=Ласкаво просимо до NeoMind
uk.WelcomeText=Нотатки й код в одному швидкому локальному просторі. Виберіть мову та набори — усе залишається на цьому комп'ютері.
uk.TypeFull=Повне встановлення (застосунок і всі набори)
uk.TypeCompact=Компактне (лише застосунок)
uk.TypeCustom=Вибіркове встановлення
uk.CompCore=Застосунок NeoMind (обов'язково)
uk.CompWeb=Веброзробка — стартові файли HTML, CSS і JavaScript
uk.CompGame=Геймдев — невеликий ігровий цикл на canvas
uk.CompCyber=Кібербезпека — безпечний скрипт розвідки й шаблон нотаток
uk.CompData=Дані та ШІ — аналіз CSV без залежностей
uk.CompEditor=Редактор і продуктивність — сніпети й приклад розкладки
uk.AddToPath=Додати NeoMind до PATH користувача (CLI)

es.WelcomeTitle=Bienvenido a NeoMind
es.WelcomeText=Tus notas y tu código en un espacio local y rápido. Elige un idioma y los paquetes que quieras: todo se queda en este equipo.
es.TypeFull=Instalación completa (app y todos los paquetes)
es.TypeCompact=Compacta (solo la app)
es.TypeCustom=Selección personalizada
es.CompCore=Aplicación NeoMind (obligatoria)
es.CompWeb=Desarrollo web: archivos iniciales de HTML, CSS y JavaScript
es.CompGame=Desarrollo de juegos: un pequeño bucle de juego en canvas
es.CompCyber=Ciberseguridad: script de reconocimiento seguro y plantilla de notas
es.CompData=Datos e IA: análisis de CSV sin dependencias
es.CompEditor=Editor y productividad: fragmentos y un mapa de teclas de ejemplo
es.AddToPath=Añadir NeoMind al PATH del usuario (CLI)

de.WelcomeTitle=Willkommen bei NeoMind
de.WelcomeText=Notizen und Code in einem schnellen lokalen Arbeitsbereich. Wähle eine Sprache und die gewünschten Pakete — alles bleibt auf diesem PC.
de.TypeFull=Vollständige Installation (App und alle Pakete)
de.TypeCompact=Kompakt (nur die App)
de.TypeCustom=Benutzerdefinierte Auswahl
de.CompCore=NeoMind-Anwendung (erforderlich)
de.CompWeb=Webentwicklung — Startdateien für HTML, CSS und JavaScript
de.CompGame=Spieleentwicklung — eine kleine Canvas-Spielschleife
de.CompCyber=Cybersicherheit — ein sicheres Recon-Skript und eine Notizvorlage
de.CompData=Daten und KI — CSV-Analyse ohne Abhängigkeiten
de.CompEditor=Editor und Produktivität — Snippets und eine Beispiel-Tastenbelegung
de.AddToPath=NeoMind zum Benutzer-PATH hinzufügen (CLI)

fr.WelcomeTitle=Bienvenue dans NeoMind
fr.WelcomeText=Vos notes et votre code dans un espace local rapide. Choisissez une langue et les packs souhaités — tout reste sur cet ordinateur.
fr.TypeFull=Installation complète (application et tous les packs)
fr.TypeCompact=Compacte (application seule)
fr.TypeCustom=Sélection personnalisée
fr.CompCore=Application NeoMind (obligatoire)
fr.CompWeb=Développement web — fichiers de départ HTML, CSS et JavaScript
fr.CompGame=Développement de jeux — une petite boucle de jeu en canvas
fr.CompCyber=Cybersécurité — script de reconnaissance sûr et modèle de notes
fr.CompData=Données et IA — analyse CSV sans dépendances
fr.CompEditor=Éditeur et productivité — extraits et exemple de raccourcis
fr.AddToPath=Ajouter NeoMind au PATH de l'utilisateur (CLI)

zh.WelcomeTitle=欢迎使用 NeoMind
zh.WelcomeText=笔记与代码，尽在一个快速的本地工作区。选择语言和需要的入门包，所有内容都保存在这台电脑上。
zh.TypeFull=完整安装（应用和所有入门包）
zh.TypeCompact=精简安装（仅应用）
zh.TypeCustom=自定义选择
zh.CompCore=NeoMind 应用（必需）
zh.CompWeb=Web 开发 — HTML、CSS 和 JavaScript 入门文件
zh.CompGame=游戏开发 — 一个小型 canvas 游戏循环
zh.CompCyber=网络安全 — 安全的侦察脚本和笔记模板
zh.CompData=数据与 AI — 无依赖的 CSV 分析入门
zh.CompEditor=编辑器与效率 — 代码片段和示例快捷键
zh.AddToPath=将 NeoMind 添加到用户 PATH（CLI）

ja.WelcomeTitle=NeoMind へようこそ
ja.WelcomeText=メモとコードを 1 つの高速なローカル作業スペースに。言語と必要なスターターパックを選んでください。すべてこの PC に保存されます。
ja.TypeFull=フルインストール（アプリとすべてのパック）
ja.TypeCompact=コンパクト（アプリのみ）
ja.TypeCustom=カスタム選択
ja.CompCore=NeoMind アプリ（必須）
ja.CompWeb=Web 開発 — HTML、CSS、JavaScript のスターターファイル
ja.CompGame=ゲーム開発 — 小さな canvas ゲームループ
ja.CompCyber=サイバーセキュリティ — 安全な偵察スクリプトとメモテンプレート
ja.CompData=データと AI — 依存関係なしの CSV 分析スターター
ja.CompEditor=エディターと生産性 — スニペットとキーマップ例
ja.AddToPath=NeoMind をユーザー PATH に追加（CLI）

[Types]
Name: "full"; Description: "{cm:TypeFull}"
Name: "compact"; Description: "{cm:TypeCompact}"
Name: "custom"; Description: "{cm:TypeCustom}"; Flags: iscustom

[Components]
Name: "core"; Description: "{cm:CompCore}"; Types: full compact custom; Flags: fixed
Name: "web"; Description: "{cm:CompWeb}"; Types: full custom
Name: "gamedev"; Description: "{cm:CompGame}"; Types: full custom
Name: "cyber"; Description: "{cm:CompCyber}"; Types: full custom
Name: "data"; Description: "{cm:CompData}"; Types: full custom
Name: "editor"; Description: "{cm:CompEditor}"; Types: full custom

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"
Name: "addtopath"; Description: "{cm:AddToPath}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
Source: "..\dist\neomind.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\dist\neomind-gui.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\dist\neomind-core.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\assets\neomind.ico"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\extras\packs\web\*"; DestDir: "{%USERPROFILE}\NeoMind Projects\Web"; Components: web; Flags: recursesubdirs createallsubdirs uninsneveruninstall
Source: "..\extras\packs\gamedev\*"; DestDir: "{%USERPROFILE}\NeoMind Projects\GameDev"; Components: gamedev; Flags: recursesubdirs createallsubdirs uninsneveruninstall
Source: "..\extras\packs\cyber\*"; DestDir: "{%USERPROFILE}\NeoMind Projects\Cybersecurity"; Components: cyber; Flags: recursesubdirs createallsubdirs uninsneveruninstall
Source: "..\extras\packs\data\*"; DestDir: "{%USERPROFILE}\NeoMind Projects\Data"; Components: data; Flags: recursesubdirs createallsubdirs uninsneveruninstall
Source: "..\extras\packs\editor\*"; DestDir: "{%USERPROFILE}\NeoMind Projects\Editor"; Components: editor; Flags: recursesubdirs createallsubdirs uninsneveruninstall

[Icons]
Name: "{group}\NeoMind"; Filename: "{app}\{#MyAppExeName}"; WorkingDir: "{app}"; IconFilename: "{app}\neomind.ico"
Name: "{group}\Uninstall NeoMind"; Filename: "{uninstallexe}"
Name: "{autodesktop}\NeoMind"; Filename: "{app}\{#MyAppExeName}"; WorkingDir: "{app}"; IconFilename: "{app}\neomind.ico"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#MyAppName}}"; Flags: nowait postinstall skipifsilent

[Code]
procedure WritePreset(Dir, FileName, DisplayName, FolderPath: string);
var
  Json: string;
begin
  StringChangeEx(FolderPath, '\', '\\', True);
  Json := '{"name":"' + DisplayName + '","path":"' + FolderPath + '"}';
  SaveStringToFile(Dir + '\' + FileName + '.json', Json, False);
end;

procedure AddToUserPath(Dir: string);
var
  Current: string;
begin
  if not RegQueryStringValue(HKCU, 'Environment', 'Path', Current) then
    Current := '';
  if Pos(Lowercase(Dir), Lowercase(Current)) = 0 then
  begin
    if (Current <> '') and (Current[Length(Current)] <> ';') then
      Current := Current + ';';
    RegWriteExpandStringValue(HKCU, 'Environment', 'Path', Current + Dir);
  end;
end;

procedure RemoveFromUserPath(Dir: string);
var
  Current, Updated: string;
begin
  if RegQueryStringValue(HKCU, 'Environment', 'Path', Current) then
  begin
    Updated := Current;
    StringChangeEx(Updated, ';' + Dir, '', True);
    StringChangeEx(Updated, Dir + ';', '', True);
    if Updated = Dir then
      Updated := '';
    if Updated <> Current then
      RegWriteExpandStringValue(HKCU, 'Environment', 'Path', Updated);
  end;
end;

procedure InitializeWizard();
begin
  WizardForm.WelcomeLabel1.Caption := ExpandConstant('{cm:WelcomeTitle}');
  WizardForm.WelcomeLabel2.Caption := ExpandConstant('{cm:WelcomeText}');
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  Home, Presets, Projects: string;
begin
  if CurStep = ssPostInstall then
  begin
    Home := ExpandConstant('{%USERPROFILE}\.neomind');
    ForceDirectories(Home);
    SaveStringToFile(Home + '\locale', ActiveLanguage(), False);

    Presets := Home + '\presets';
    DelTree(Presets, True, True, True);
    ForceDirectories(Presets);
    Projects := ExpandConstant('{%USERPROFILE}\NeoMind Projects');

    if WizardIsComponentSelected('web') then
      WritePreset(Presets, 'web', 'Web', Projects + '\Web');
    if WizardIsComponentSelected('gamedev') then
      WritePreset(Presets, 'gamedev', 'GameDev', Projects + '\GameDev');
    if WizardIsComponentSelected('cyber') then
      WritePreset(Presets, 'cyber', 'Cybersecurity', Projects + '\Cybersecurity');
    if WizardIsComponentSelected('data') then
      WritePreset(Presets, 'data', 'Data', Projects + '\Data');
    if WizardIsComponentSelected('editor') then
      WritePreset(Presets, 'editor', 'Editor', Projects + '\Editor');

    if WizardIsTaskSelected('addtopath') then
      AddToUserPath(ExpandConstant('{app}'));
  end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usPostUninstall then
    RemoveFromUserPath(ExpandConstant('{app}'));
end;
