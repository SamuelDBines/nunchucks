// const package_json_path = '../package.json'
const packageJson = require('../package.json')
const { execSync } = require('node:child_process')
const fs = require('node:fs')

const oldtag = packageJson.version;
const main = () => {
  try {
    const args = process.argv.slice(2);

    const versionArg = args[0];
    let versionStr = ''
    let tag = packageJson.version;
    const packageJsonVersion = packageJson.version.split('.')
    packageJsonVersion[3] = undefined
    if(versionArg) {
      
      if(versionArg.toLocaleLowerCase() == 'patch') {
        versionStr = 'patch'
        packageJsonVersion[2] = Number(packageJsonVersion[2]) + 1 
      }
      if(versionArg.toLocaleLowerCase() == 'minor') {
        versionStr = 'minor'
        packageJsonVersion[1] = Number(packageJsonVersion[1]) + 1 
        packageJsonVersion[2] = '0'
        packageJsonVersion[3] = undefined
      }
      if(versionArg.toLocaleLowerCase() == 'major') {
        versionStr = 'major'
        packageJsonVersion[0] = Number(packageJsonVersion[0]) + 1 
        packageJsonVersion[1] = '0'
        packageJsonVersion[2] = '0'
        
      }
    }
    tag = packageJsonVersion.filter(Boolean).join('.') 
    packageJson.version = tag
    fs.writeFileSync('package.json', JSON.stringify(packageJson, undefined, 2))
    execSync(`git add .`);
    execSync(`git commit -m "release latest version"`);
    try {
      execSync(`~/scripts/shared/tag ${tag}`)
    } catch(err) {
      console.log('Error tagging: ', err)
      execSync(`git push`);
    }
  } catch(err) {
    //Revert old tag
    packageJson.version = oldtag
    fs.writeFileSync('package.json', JSON.stringify(packageJson, undefined, 2))
    console.error(err);
  }
  // console.log(JSON.stringify(packageJson, undefined, 2))
}

main()

// pnpm publish --access=public --no-git-checks