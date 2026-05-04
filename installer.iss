#ifndef AppVersion
#define AppVersion "0.0.0-dev"
#endif

[Setup]
AppName=searchdisk
AppVersion={#AppVersion}
AppPublisher=shuntaka9576
DefaultDirName={autopf}\searchdisk
DefaultGroupName=searchdisk
OutputDir=dist
OutputBaseFilename=searchdisk-setup
Compression=lzma2
SolidCompression=yes
ChangesEnvironment=yes
ArchitecturesInstallIn64BitMode=x64compatible
PrivilegesRequired=lowest

[Files]
Source: "dist\searchdisk.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "README-windows.txt"; DestDir: "{app}"; Flags: ignoreversion

[Registry]
Root: HKCU; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: NeedsAddPath(ExpandConstant('{app}'))

[Code]
function NeedsAddPath(Param: string): Boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', OrigPath) then
  begin
    Result := True;
    exit;
  end;
  Result := Pos(';' + Param + ';', ';' + OrigPath + ';') = 0;
end;
